package models

import(
	"github.com/google/uuid"
	"time"
)

type MfaAmrClaims struct {
	SessionId uuid.UUID `json:"session_id,omitempty" column:"session_id"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" column:"updated_at"`
	AuthenticationMethod string `json:"authentication_method,omitempty" column:"authentication_method"`
	Id uuid.UUID `json:"id,omitempty" column:"id"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
