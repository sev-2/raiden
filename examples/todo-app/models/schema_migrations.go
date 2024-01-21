package models

type SchemaMigrations struct {
	Version string `json:"version,omitempty" column:"version"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
