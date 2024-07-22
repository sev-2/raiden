package db

import "time"

type Foo struct {
	ModelBase

	Id        int64     `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable:false;default:now()"`
	Body      string    `json:"body,omitempty" column:"name:body;type:text;nullable:false"`

	Metadata string `json:"-" schema:"public" tableName:"foo" rlsEnable:"true" rlsForced:"false"`

	Acl string `json:"-" read:"" write:""`
}
