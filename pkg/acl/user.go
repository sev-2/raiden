package acl

import (
	"net/http"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/jwt"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

var Logger hclog.Logger = logger.HcLog().Named("acl")

const ServiceRoleName = "service_role"

func SetUserRole(cfg *raiden.Config, userId string, role raiden.Role) error {
	if err := validateRole(role.Name()); err != nil {
		return err
	}
	return doSetRole(cfg, userId, role.Name())
}

func doSetRole(cfg *raiden.Config, userId string, roleName string) error {
	// call supabase with admin privilege for update user, user service account for this case
	data := objects.User{Role: roleName}
	_, err := supabase.AdminUpdateUserData(cfg, userId, data)
	if err != nil {
		return err
	}
	return nil
}

func Authenticated(ctx raiden.Context) error {
	authHeader := string(ctx.RequestContext().Request.Header.Peek("Authorization"))

	if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return &raiden.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Code:       "unauthorize",
			Message:    "unauthorize - invalid format",
		}
	}

	token := strings.TrimPrefix(strings.ToLower(authHeader), "bearer ")
	if len(token) == 0 {
		return &raiden.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Code:       "unauthorize",
			Message:    "unauthorize - required token",
		}
	}

	_, err := jwt.Validate[jwt.JWTClaims](token, ctx.Config().JwtSecret)
	if err != nil {
		Logger.Error("validation failed", "message", err)
		return &raiden.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Code:       "unauthorize",
			Message:    "unauthorize - invalid token",
		}
	}

	return nil
}

func GetAuthenticatedData(ctx raiden.Context, serviceOnly bool) (*jwt.JWTClaims, error) {
	authHeader := string(ctx.RequestContext().Request.Header.Peek("Authorization"))

	if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return nil, &raiden.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Code:       "unauthorize",
			Message:    "unauthorize - invalid format",
		}
	}

	token := strings.TrimSpace(authHeader)
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimPrefix(token, "bearer ")

	if len(token) == 0 {
		return nil, &raiden.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Code:       "unauthorize",
			Message:    "unauthorize - required token",
		}
	}

	data, err := jwt.Validate[jwt.JWTClaims](token, ctx.Config().JwtSecret)
	if err != nil {
		Logger.Error("validation failed", "message", err)
		return nil, &raiden.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Code:       "unauthorize",
			Message:    "unauthorize - invalid token",
		}
	}

	if serviceOnly && data.Role != ServiceRoleName {
		Logger.Error("invalid role", "message", "access for service only")
		return nil, &raiden.ErrorResponse{
			StatusCode: http.StatusForbidden,
			Code:       "forbidden",
			Message:    "You do not have permission to access this resource.",
		}
	}

	return data, nil
}
