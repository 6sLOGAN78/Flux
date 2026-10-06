package email

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/templates"
	"github.com/resend/resend-go/v2"
	"github.com/rs/zerolog"
)

type Client struct {
	client *resend.Client
	logger *zerolog.Logger
}

func NewClient(cfg *config.Config, logger *zerolog.Logger) *Client {
	return &Client{
		client: resend.NewClient(cfg.Integration.ResendAPIKey),
		logger: logger,
	}
}

// Render formats an embedded template without contacting the email provider.
func (c *Client) Render(templateName Template, data map[string]string) (string, error) {
	var tmplPath string
	switch templateName {
	case TemplateWelcome:
		tmplPath = "emails/welcome.html"
	default:
		return "", fmt.Errorf("unsupported email template")
	}

	tmpl, err := template.New("welcome.html").Option("missingkey=error").ParseFS(templates.Assets, tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse email template %s: %w", templateName, err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return "", fmt.Errorf("failed to execute email template %s: %w", templateName, err)
	}
	return body.String(), nil
}

func (c *Client) SendEmail(to, subject string, templateName Template, data map[string]string) error {
	body, err := c.Render(templateName, data)
	if err != nil {
		return err
	}
	if c == nil || c.client == nil {
		return fmt.Errorf("email transport is not configured")
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
