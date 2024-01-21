package models

import(
	"github.com/google/uuid"
	"time"
)

type SsoProviders struct {
	Id uuid.UUID `json:"id,omitempty" column:"id"`
	ResourceId *string `json:"resource_id,omitempty" column:"resource_id"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" column:"updated_at"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
