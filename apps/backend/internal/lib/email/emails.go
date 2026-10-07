package email

// SendWelcomeEmail renders and sends the embedded welcome template.
func (c *Client) SendWelcomeEmail(to, firstName string) error {
	data := map[string]string{
		"UserFirstName": firstName,
	}

	return c.SendEmail(
		to,
		"Welcome to Flux!",
		TemplateWelcome,
		data,
	)
}
