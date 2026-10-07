package email

// PreviewData returns a fresh set of example values for local email previews.
func PreviewData() map[string]map[string]string {
	return map[string]map[string]string{
		"welcome": {
			"UserFirstName": "John",
		},
	}
}
