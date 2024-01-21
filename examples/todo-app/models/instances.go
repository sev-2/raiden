package models

import(
	"github.com/google/uuid"
	"time"
)

type Instances struct {
	Id uuid.UUID `json:"id,omitempty" column:"id"`
	Uuid *uuid.UUID `json:"uuid,omitempty" column:"uuid"`
	RawBaseConfig *string `json:"raw_base_config,omitempty" column:"raw_base_config"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" column:"updated_at"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
