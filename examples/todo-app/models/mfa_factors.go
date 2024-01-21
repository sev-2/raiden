package models

import(
	"github.com/google/uuid"
	"time"
)

type MfaFactors struct {
	Id uuid.UUID `json:"id,omitempty" column:"id"`
	UserId uuid.UUID `json:"user_id,omitempty" column:"user_id"`
	FriendlyName *string `json:"friendly_name,omitempty" column:"friendly_name"`
	FactorType any `json:"factor_type,omitempty" column:"factor_type"`
	Status any `json:"status,omitempty" column:"status"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" column:"updated_at"`
	Secret *string `json:"secret,omitempty" column:"secret"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
