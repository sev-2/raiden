package models

import(
	"github.com/google/uuid"
	"time"
)

type SamlRelayStates struct {
	Id uuid.UUID `json:"id,omitempty" column:"id"`
	SsoProviderId uuid.UUID `json:"sso_provider_id,omitempty" column:"sso_provider_id"`
	RequestId string `json:"request_id,omitempty" column:"request_id"`
	ForEmail *string `json:"for_email,omitempty" column:"for_email"`
	RedirectTo *string `json:"redirect_to,omitempty" column:"redirect_to"`
	FromIpAddress *any `json:"from_ip_address,omitempty" column:"from_ip_address"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" column:"updated_at"`
	FlowStateId *uuid.UUID `json:"flow_state_id,omitempty" column:"flow_state_id"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
