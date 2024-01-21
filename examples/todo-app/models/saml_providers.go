package models

import(
	"github.com/google/uuid"
	"encoding/json"
	"time"
)

type SamlProviders struct {
	Id uuid.UUID `json:"id,omitempty" column:"id"`
	SsoProviderId uuid.UUID `json:"sso_provider_id,omitempty" column:"sso_provider_id"`
	EntityId string `json:"entity_id,omitempty" column:"entity_id"`
	MetadataXml string `json:"metadata_xml,omitempty" column:"metadata_xml"`
	MetadataUrl *string `json:"metadata_url,omitempty" column:"metadata_url"`
	AttributeMapping *json.RawMessage `json:"attribute_mapping,omitempty" column:"attribute_mapping"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" column:"updated_at"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
