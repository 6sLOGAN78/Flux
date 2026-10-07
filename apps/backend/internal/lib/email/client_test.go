package email_test

import (
	"strings"
	"testing"

	"github.com/6sLOGAN78/flux/internal/lib/email"
)

func TestRenderAlternateDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	client := &email.Client{}
	input := `<script>alert("x")</script> & 'visitor'`
	body, err := client.Render(email.TemplateWelcome, map[string]string{"UserFirstName": input})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(body, input) || strings.Contains(body, "<script>") {
		t.Fatal("untrusted name rendered as HTML")
	}
	if !strings.Contains(body, "&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt; &amp; &#39;visitor&#39;") {
		t.Fatal("name was not escaped and interpolated")
	}
	if strings.Contains(body, "{{.UserFirstName}}") {
		t.Fatal("Go token was not executed")
	}
	if !strings.Contains(body, "Flux. All rights reserved.") || strings.Contains(body, "Alfred") {
		t.Fatal("welcome footer must use the stable Flux name")
	}
	bodyAgain, err := client.Render(email.TemplateWelcome, map[string]string{"UserFirstName": input})
	if err != nil || bodyAgain != body {
		t.Fatalf("repeated rendering differs: %v", err)
	}
}

func TestTemplateUnknownRejectedBeforeSend(t *testing.T) {
	// No transport is configured: invalid names must fail before touching it.
	client := &email.Client{}
	for _, name := range []email.Template{"",
		"unknown",
		"../welcome",
		"../../static/openapi",
		"/tmp/welcome",
		"welcome.html",
		"emails/welcome",
		"WELCOME"} {
		t.Run(string(name), func(t *testing.T) {
			body, err := client.Render(name, nil)
			if err == nil || body != "" {
				t.Fatal("unknown template rendered successfully")
			}
			if err45 := client.SendEmail("visitor@example.com", "Welcome", name, nil); err45 == nil {
				t.Fatal("unknown template reached transport")
			}
		})
	}
}

func TestRenderMissingRequiredData(t *testing.T) {
	client := &email.Client{}
	if body, err := client.Render(email.TemplateWelcome, nil); err == nil || body != "" {
		t.Fatal("missing template data must not render a broken email")
	}
}
