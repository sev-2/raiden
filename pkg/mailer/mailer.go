package mailer

import (
	"fmt"
	"net/url"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/supabase/mailme"
	"gopkg.in/gomail.v2"
)

var MailLogger = logger.HcLog().Named("mailer")

type Mailer interface {
	RecoveryMail(email, token, otp, referrerURL string, externalURL *url.URL) error
}

type EmailParams struct {
	Token      string
	Type       string
	RedirectTo string
}

func NewMailer(config *raiden.Config) Mailer {

	mail := gomail.NewMessage()

	from := mail.FormatAddress(config.SmtpAdminEmail, config.SmtpSenderName)

	var mailClient = &mailme.Mailer{
		Host:    config.SmtpHost,
		Port:    config.SmtpPort,
		User:    config.SmtpUser,
		Pass:    config.SmtpPass,
		From:    from,
		BaseURL: config.SupabasePublicUrl,
	}

	return &TemplateMailer{
		SiteURL: config.SupabasePublicUrl,
		Config:  config,
		Mailer:  mailClient,
	}
}

func getPath(filepath string, params *EmailParams) (*url.URL, error) {
	path := &url.URL{}
	if filepath != "" {
		if p, err := url.Parse(filepath); err != nil {
			return nil, err
		} else {
			path = p
		}
	}
	if params != nil {
		path.RawQuery = fmt.Sprintf("token=%s&type=%s&redirect_to=%s", url.QueryEscape(params.Token), url.QueryEscape(params.Type), encodeRedirectURL(params.RedirectTo))
	}
	return path, nil
}
