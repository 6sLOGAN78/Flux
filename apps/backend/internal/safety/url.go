// Package safety validates destinations without DNS or HTTP egress.
package safety

import (
	"errors"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/6sLOGAN78/flux/internal/config"
)

// ValidateDestination returns a canonical public HTTP(S) URL without resolving or fetching it.
func ValidateDestination(value string, policy config.LinksConfig) (string, error) {
	invalid := errors.New("invalid destination")
	u, err := parseDestination(value)
	if err != nil {
		return "", err
	}
	if policy.ManagedHost == "" {
		return "", invalid
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
	if strings.HasPrefix(u.Host, "[") {
		if address, ipErr := netip.ParseAddr(u.Hostname()); ipErr != nil || !address.Is6() {
			return nil, invalid
		}
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
	// IPv6 global unicast is currently allocated only within 2000::/3.
	// Go IsGlobalUnicast includes unallocated and deprecated site-local space.
	if address.Is6() && !netip.MustParsePrefix("2000::/3").Contains(address) {
		return false
	}
	for _, cidr := range []string{
		"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24", "198.18.0.0/15", "198.51.100.0/24",
		"203.0.113.0/24", "240.0.0.0/4", "192.88.99.0/24",
		"192.31.196.0/24", "192.52.193.0/24", "192.175.48.0/24", "2620:4f:8000::/48",
		"2001::/23", "2001:db8::/32", "2002::/16", "3fff::/20",
		"64:ff9b::/96", "64:ff9b:1::/48",
	} {
		if netip.MustParsePrefix(cidr).Contains(address) {
			return false
		}
	}
	return true
}
