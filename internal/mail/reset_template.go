package mail

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
)

//go:embed ResetPasswordTemplate.html
var resetPasswordTemplateHTML string

// ResetPasswordEmailData is passed to the HTML template for password reset emails.
type ResetPasswordEmailData struct {
	Link     string
	SiteName string
}

// RenderResetPasswordHTML returns the rendered HTML body for a reset-password email.
func RenderResetPasswordHTML(data ResetPasswordEmailData) (string, error) {
	tmpl, err := template.New("reset_password").Parse(resetPasswordTemplateHTML)
	if err != nil {
		return "", fmt.Errorf("mail: parse reset template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("mail: execute reset template: %w", err)
	}
	return buf.String(), nil
}
