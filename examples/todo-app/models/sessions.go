package models

import(
	"github.com/google/uuid"
	"time"
)

type Sessions struct {
	Id uuid.UUID `json:"id,omitempty" column:"id"`
	UserId uuid.UUID `json:"user_id,omitempty" column:"user_id"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" column:"updated_at"`
	FactorId *uuid.UUID `json:"factor_id,omitempty" column:"factor_id"`
	Aal *any `json:"aal,omitempty" column:"aal"`
	NotAfter *time.Time `json:"not_after,omitempty" column:"not_after"`
	RefreshedAt *time.Time `json:"refreshed_at,omitempty" column:"refreshed_at"`
	UserAgent *string `json:"user_agent,omitempty" column:"user_agent"`
	Ip *any `json:"ip,omitempty" column:"ip"`
	Tag *string `json:"tag,omitempty" column:"tag"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
