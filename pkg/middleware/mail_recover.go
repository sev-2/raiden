package auth

import (
	"net/url"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/crypto"
	"github.com/sev-2/raiden/pkg/mailer"
	"github.com/sev-2/raiden/pkg/supabase"
)

func sendPasswordRecovery(ctx raiden.Context) error {
	var err error

	mailer := mailer.NewMailer(ctx.Config())

	email := ctx.Get("email").(string)
	user, err := supabase.GetUserByEmail(ctx.Config(), email)

	if err != nil {
		return err
	}

	otp, err := crypto.GenerateOtp(6)

	if err != nil {
		return err
	}

	token := crypto.GenerateTokenHash(user.Email, otp)

	supabase.UpdateUserByEmail(ctx.Config(), email, token)
	referrerURL := ctx.Get("redirect_to").(string)
	externalURL := ctx.Get("externalURL").(string)

	extUrl, err := url.Parse(externalURL)

	if (err != nil) || (extUrl.Scheme == "") || (extUrl.Host == "") {
		return err
	}

	if err := mailer.RecoveryMail(email, token, otp, referrerURL, extUrl); err != nil {
		return err
	}

	return err
}

func PasswordRecoveryMiddleware(path string) raiden.MiddlewareFn {
	return func(next raiden.RouteHandlerFn) raiden.RouteHandlerFn {
		return func(ctx raiden.Context) error {

			if ctx.Config().Mode == raiden.BffMode {

				if path != "/auth/v1/recover" {
					return next(ctx)
				}

				err := sendPasswordRecovery(ctx)
				if err != nil {
					return err
				}

				return ctx.SendJson(map[string]string{"message": "success", "code": "OK"})
			}

			return next(ctx)

		}
	}
}
