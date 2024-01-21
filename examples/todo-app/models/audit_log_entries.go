package models

import(
	"encoding/json"
	"time"
	"github.com/google/uuid"
)

type AuditLogEntries struct {
	InstanceId *uuid.UUID `json:"instance_id,omitempty" column:"instance_id"`
	Id uuid.UUID `json:"id,omitempty" column:"id"`
	Payload *json.RawMessage `json:"payload,omitempty" column:"payload"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"created_at"`
	IpAddress string `json:"ip_address,omitempty" column:"ip_address"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
