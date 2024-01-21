package models

import(
	"github.com/google/uuid"
	"encoding/json"
	"time"
)

type Identities struct {
	ProviderId string `json:"provider_id,omitempty" column:"provider_id"`
	UserId uuid.UUID `json:"user_id,omitempty" column:"user_id"`
	IdentityData json.RawMessage `json:"identity_data,omitempty" column:"identity_data"`
	Provider string `json:"provider,omitempty" column:"provider"`
	LastSignInAt *time.Time `json:"last_sign_in_at,omitempty" column:"last_sign_in_at"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" column:"updated_at"`
	Email *string `json:"email,omitempty" column:"email"`
	Id uuid.UUID `json:"id,omitempty" column:"id"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
