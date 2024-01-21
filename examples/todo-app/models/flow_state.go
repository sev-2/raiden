package models

import(
	"github.com/google/uuid"
	"time"
)

type FlowState struct {
	Id uuid.UUID `json:"id,omitempty" column:"id"`
	UserId *uuid.UUID `json:"user_id,omitempty" column:"user_id"`
	AuthCode string `json:"auth_code,omitempty" column:"auth_code"`
	CodeChallengeMethod any `json:"code_challenge_method,omitempty" column:"code_challenge_method"`
	CodeChallenge string `json:"code_challenge,omitempty" column:"code_challenge"`
	ProviderType string `json:"provider_type,omitempty" column:"provider_type"`
	ProviderAccessToken *string `json:"provider_access_token,omitempty" column:"provider_access_token"`
	ProviderRefreshToken *string `json:"provider_refresh_token,omitempty" column:"provider_refresh_token"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" column:"updated_at"`
	AuthenticationMethod string `json:"authentication_method,omitempty" column:"authentication_method"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
