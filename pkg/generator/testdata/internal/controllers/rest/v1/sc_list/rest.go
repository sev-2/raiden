package sclist

import (
	"time"

	"github.com/sev-2/raiden"
)

type Sc struct {
	raiden.ModelBase
	Id          int64      `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	Name        string     `json:"name,omitempty" column:"name:name;type:varchar;nullable:false"`
	Description *string    `json:"description,omitempty" column:"name:description;type:text;nullable"`
	IsActive    *bool      `json:"is_active,omitempty" column:"name:is_active;type:boolean;nullable;default:true"`
	CreatedAt   *time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable;default:now()"`

	// Table information
	Metadata string `json:"-" schema:"public" rlsEnable:"true" rlsForced:"false"`

	// Access control
	Acl string `json:"-" read:"authenticated,sc" write:"authenticated,sc" readUsing:"true" writeCheck:"true" writeUsing:"true"`
}

type ScController struct {
	raiden.ControllerBase
	Model Sc
}

func (sc *ScController) BeforeGet(ctx raiden.Context) error {
	raiden.Info("[ScController] before get fire")
	return nil
}

func (sc *ScController) AfterGet(ctx raiden.Context) error {
	raiden.Info("[ScController] after get fire")
	return nil
}
