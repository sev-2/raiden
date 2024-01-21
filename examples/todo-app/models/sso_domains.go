package models

import(
	"github.com/google/uuid"
	"time"
)

type SsoDomains struct {
	Id uuid.UUID `json:"id,omitempty" column:"id"`
	SsoProviderId uuid.UUID `json:"sso_provider_id,omitempty" column:"sso_provider_id"`
	Domain string `json:"domain,omitempty" column:"domain"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" column:"updated_at"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
