package models

import(
	"time"
)

type Tasks struct {
	Id int64 `json:"id,omitempty" column:"id"`
	Name *string `json:"name,omitempty" column:"name"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"created_at"`

	Metadata string `schema:"public"`
	Acl string `read:"" write:"authenticated"`
}
