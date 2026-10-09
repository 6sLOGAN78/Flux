package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LinkService enforces destination policy and fresh server-side capabilities.
type LinkService struct {
	store     *repository.LinkRepository
	workspace *WorkspaceService
	policy    config.LinksConfig
}

// NewLinkService injects transactional storage and validated operator policy.
func NewLinkService(store *repository.LinkRepository, workspace *WorkspaceService,
	policy config.LinksConfig,
) *LinkService {
	return &LinkService{store: store, workspace: workspace, policy: policy}
}

// ManagedHost exposes only the fixed operator-configured management hostname.
func (s *LinkService) ManagedHost() string { return s.policy.ManagedHost }

// Create validates canonical payload without DNS, HTTP or preview egress.
func (s *LinkService) Create(ctx context.Context, scope repository.Scope,
	destination, title, key string,
) (repository.Link, error) {
	canonical, err := validateDestination(destination, s.policy)
	keyOK, matchErr := regexp.MatchString(`^[A-Za-z0-9_-]{16,128}$`, key)
	if err != nil || matchErr != nil || !keyOK || !utf8.ValidString(title) || utf8.RuneCountInString(title) > 200 {
		return repository.Link{}, errs.NewBadRequestError("Enter a public HTTP(S) destination, "+
			"a title of at most 200 characters and a valid retry key.", false, nil, nil, nil)
	}
	payload, err := json.Marshal(struct {
		Destination string `json:"destination"`
		Title       string `json:"title"`
	}{canonical, title})
	if err != nil {
		return repository.Link{}, linkFailure(err)
	}
	hash := sha256.Sum256(payload)
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	result, err := s.store.Create(ctx, scope, s.policy.ManagedHost, canonical, title,
		key, hash[:], s.workspace.RequireWrite)
	return result, linkFailure(err)
}

// Detail preserves durable creator provenance without granting former access.
func (s *LinkService) Detail(ctx context.Context, scope repository.Scope, id uuid.UUID) (repository.Link, error) {
	ctx, cancel := context.WithTimeout(ctx, workspaceTimeout)
	defer cancel()
	authorize := func(ctx context.Context, tx pgx.Tx, scope repository.Scope) error {
		return s.workspace.RequireCapability(ctx, tx, scope, CapabilityRead)
	}
	result, err := s.store.Detail(ctx, scope, id, authorize)
	return result, linkFailure(err)
}

func linkFailure(err error) error {
	if err == nil {
		return nil
	}
	var typed *errs.HTTPError
	if errors.As(err, &typed) {
		return err
	}
	if errors.Is(err, repository.ErrWorkspaceConflict) {
		return &errs.HTTPError{Code: "IDEMPOTENCY_CONFLICT",
			Message: "This retry key was used for a different request.", Status: http.StatusConflict}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.NewNotFoundError("Link not found", false, nil)
	}
	return errors.Join(&errs.HTTPError{Code: errs.MakeUpperCaseWithUnderscores(
		http.StatusText(http.StatusServiceUnavailable)), Message: "Links temporarily unavailable",
		Status: http.StatusServiceUnavailable}, err)
}

func validateDestination(value string, policy config.LinksConfig) (string, error) {
	invalid := errors.New("invalid destination")
	u, err := parseDestination(value)
	if err != nil {
		return "", err
	}
	host, err := publicHostname(u.Hostname())
	if err != nil {
		return "", err
	}
	blockedHosts := append([]string{policy.ManagedHost, "localhost", "local", "internal",
		"metadata.google.internal", "metadata.google", "instance-data.ec2.internal"}, policy.BlockedHosts...)
	for _, blocked := range blockedHosts {
		if host == blocked || strings.HasSuffix(host, "."+blocked) {
			return "", invalid
		}
	}
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	if port := u.Port(); port != "" {
		host += ":" + port
	}
	u.Host = host
	return u.String(), nil
}

func destinationTextOK(value string) bool {
	if !utf8.ValidString(value) || len(value) == 0 || len(value) > 8192 || strings.ContainsRune(value, '\\') {
		return false
	}
	for _, c := range value {
		if unicode.IsControl(c) || unicode.IsSpace(c) {
			return false
		}
	}
	decoded, decodeErr := url.PathUnescape(value)
	if decodeErr != nil || strings.ContainsRune(decoded, '\\') {
		return false
	}
	for _, c := range decoded {
		if unicode.IsControl(c) {
			return false
		}
	}
	return true
}

const linkSecureScheme = "https"

func parseDestination(value string) (*url.URL, error) {
	invalid := errors.New("invalid destination")
	if !destinationTextOK(value) {
		return nil, invalid
	}
	u, err := url.Parse(value)
	if err != nil || u.Opaque != "" || u.User != nil || u.Host == "" ||
		(u.Scheme != "http" && u.Scheme != linkSecureScheme) {
		return nil, invalid
	}
	if strings.Contains(u.Host, "%") || strings.HasSuffix(u.Host, ":") {
		return nil, invalid
	}
	if port := u.Port(); port != "" {
		number, portErr := strconv.Atoi(port)
		if portErr != nil || number < 1 || number > 65535 {
			return nil, invalid
		}
	}
	return u, nil
}

func publicHostname(value string) (string, error) {
	host := strings.ToLower(strings.TrimSuffix(value, "."))
	if address, ipErr := netip.ParseAddr(host); ipErr == nil {
		address = address.Unmap()
		if !publicAddress(address) {
			return "", errors.New("invalid destination address")
		}
		return address.String(), nil
	}
	return config.NormalizeLinkHost(host)
}

func publicAddress(address netip.Addr) bool {
	if !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() {
		return false
	}
	for _, cidr := range []string{
		"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24", "198.18.0.0/15", "198.51.100.0/24",
		"203.0.113.0/24", "240.0.0.0/4", "2001:db8::/32", "2001::/32", "2002::/16", "64:ff9b::/96", "64:ff9b:1::/48",
	} {
		if netip.MustParsePrefix(cidr).Contains(address) {
			return false
		}
	}
	return true
}
