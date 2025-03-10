package mailer

import (
	"net/url"
	"strings"

	"github.com/sev-2/raiden"
)

type MailClient interface {
	Mail(string, string, string, string, map[string]interface{}) error
}

// TemplateMailer will send mail and use templates from the site for easy mail styling
type TemplateMailer struct {
	SiteURL string
	Config  *raiden.Config
	Mailer  MailClient
}

func encodeRedirectURL(referrerURL string) string {
	if len(referrerURL) > 0 {
		if strings.ContainsAny(referrerURL, "&=#") {
			// if the string contains &, = or # it has not been URL
			// encoded by the caller, which means it should be URL
			// encoded by us otherwise, it should be taken as-is
			referrerURL = url.QueryEscape(referrerURL)
		}
	}
	return referrerURL
}

const defaultRecoveryMail = `<h2>Reset password</h2>

<p>Follow this link to reset the password for your user:</p>
<p><a href="{{ .ConfirmationURL }}">Reset password</a></p>
<p>Alternatively, enter the code: {{ .Token }}</p>`

// RecoveryMail sends a password recovery mail
func (m *TemplateMailer) RecoveryMail(email string, token string, otp, referrerURL string, externalURL *url.URL) error {
	path, err := getPath("/verify", &EmailParams{
		Token:      token,
		Type:       "recovery",
		RedirectTo: referrerURL,
	})
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"SiteURL":         m.Config.SupabasePublicUrl,
		"ConfirmationURL": externalURL.ResolveReference(path).String(),
		"Email":           email,
		"Token":           otp,
		"TokenHash":       token,
		"RedirectTo":      referrerURL,
	}

	return m.Mailer.Mail(
		email,
		"Reset Your Password",
		"",
		defaultRecoveryMail,
		data,
	)
}
