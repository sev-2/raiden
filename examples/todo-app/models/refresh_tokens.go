package models

import(
	"github.com/google/uuid"
	"time"
)

type RefreshTokens struct {
	InstanceId *uuid.UUID `json:"instance_id,omitempty" column:"instance_id"`
	Id int64 `json:"id,omitempty" column:"id"`
	Token *string `json:"token,omitempty" column:"token"`
	UserId *string `json:"user_id,omitempty" column:"user_id"`
	Revoked *bool `json:"revoked,omitempty" column:"revoked"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" column:"updated_at"`
	Parent *string `json:"parent,omitempty" column:"parent"`
	SessionId *uuid.UUID `json:"session_id,omitempty" column:"session_id"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
