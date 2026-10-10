// Package email renders embedded templates and delivers transactional email.
package email

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/lib/invitationcrypto"
	"github.com/6sLOGAN78/flux/templates"
	"github.com/resend/resend-go/v2"
	"github.com/rs/zerolog"
)

// Client renders templates and sends transactional emails.
type Client struct {
	client              *resend.Client
	invitationClient    *resend.Client
	invitationTransport *http.Transport
	logger              *zerolog.Logger
	configured          bool
	closed              atomic.Bool
}

const (
	invitationSendTimeout = 10 * time.Second
	welcomeFirstNameField = "UserFirstName"
)

// NewClient constructs the configured transactional email transport.
func NewClient(cfg *config.Config, logger *zerolog.Logger) *Client {
	transport := &http.Transport{ForceAttemptHTTP2: true, TLSHandshakeTimeout: invitationSendTimeout}
	transport.Proxy = nil
	invitationClient := resend.NewCustomClient(&http.Client{Transport: transport, Timeout: invitationSendTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
		cfg.Integration.ResendAPIKey)
	invitationClient.BaseURL = &url.URL{Scheme: "https", Host: "api.resend.com", Path: "/"}
	return &Client{
		client:              resend.NewClient(cfg.Integration.ResendAPIKey),
		invitationClient:    invitationClient,
		invitationTransport: transport,
		logger:              logger,
		configured:          strings.TrimSpace(cfg.Integration.ResendAPIKey) != "",
	}
}

// Close releases the invitation adapter's owned idle transport connections.
func (c *Client) Close() error {
	if c != nil && !c.closed.Swap(true) {
		if c.invitationTransport != nil {
			c.invitationTransport.CloseIdleConnections()
		}
	}
	return nil
}

// CheckLocal validates the owned adapter and both embedded templates without
// sending mail or probing provider availability. Render remains transport-free.
func (c *Client) CheckLocal(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c == nil || c.closed.Load() || !c.configured || c.client == nil || c.invitationClient == nil {
		return errors.New("email adapter unconfigured or closed")
	}
	if _, err := c.Render(TemplateWelcome, map[string]string{welcomeFirstNameField: "Health"}); err != nil {
		return err
	}
	if _, err := c.Render(TemplateInvitation, map[string]string{
		"WorkspaceName": "Health", "Role": "member", "InvitationURL": "https://app.flux.test/invitations",
	}); err != nil {
		return err
	}
	return ctx.Err()
}

// TemplateInvitation identifies the closed embedded invitation asset.
const TemplateInvitation Template = "invitation"

// SendInvitation uses the fixed HTTPS provider with a stable intent idempotency key.
func (c *Client) SendInvitation(ctx context.Context, message invitationcrypto.Message) error {
	if c == nil || c.closed.Load() || !c.configured || c.invitationClient == nil {
		return errors.New("invitation transport unavailable")
	}
	ack, err := c.invitationClient.Emails.SendWithOptions(ctx, &resend.SendEmailRequest{
		From: message.From, To: []string{message.To}, Subject: message.Subject, Html: message.HTML},
		&resend.SendEmailOptions{IdempotencyKey: message.IdempotencyKey})
	if err != nil || ack == nil || ack.Id == "" {
		return errors.New("invitation acknowledgement unavailable")
	}
	return nil
}

// Render formats an embedded template without contacting the email provider.
func (c *Client) Render(templateName Template, data map[string]string) (string, error) {
	var tmplPath string
	switch templateName {
	case TemplateWelcome:
		tmplPath = "emails/welcome.html"
	case TemplateInvitation:
		tmplPath = "emails/invitation.html"
	default:
		return "", errors.New("unsupported email template")
	}

	tmpl, err := template.New(string(templateName)+".html").Option("missingkey=error").ParseFS(templates.Assets, tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse email template %s: %w", templateName, err)
	}

	var body bytes.Buffer
	if err46 := tmpl.Execute(&body, data); err46 != nil {
		return "", fmt.Errorf("failed to execute email template %s: %w", templateName, err46)
	}
	return body.String(), nil
}

// SendEmail delivers an email through the configured transport.
func (c *Client) SendEmail(to, subject string, templateName Template, data map[string]string) error {
	body, err := c.Render(templateName, data)
	if err != nil {
		return err
	}
	if c == nil || c.closed.Load() || !c.configured || c.client == nil {
		return errors.New("email transport is not configured")
	}

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", "Flux", "onboarding@resend.dev"),
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	_, err = c.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
