package models

import(
	"github.com/google/uuid"
	"time"
)

type MfaChallenges struct {
	Id uuid.UUID `json:"id,omitempty" column:"id"`
	FactorId uuid.UUID `json:"factor_id,omitempty" column:"factor_id"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"created_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty" column:"verified_at"`
	IpAddress any `json:"ip_address,omitempty" column:"ip_address"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
