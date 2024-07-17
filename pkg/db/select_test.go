package db

import (
	"testing"
	"time"

	"github.com/sev-2/raiden"
)

type Article struct {
	ModelBase

	Id        int64     `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	UserId    int64     `json:"user_id,omitempty" column:"name:user_id;type:bigint;nullable:false"`
	Title     string    `json:"title,omitempty" column:"name:title;type:text;nullable:false"`
	Body      string    `json:"body,omitempty" column:"name:body;type:text;nullable:true"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable:false;default:now()"`

	Metadata string `json:"-" schema:"public" tableName:"articles" rlsEnable:"true" rlsForced:"false"`
}

var model = Article{}

func TestSelect(t *testing.T) {
	ctx := raiden.Ctx{}

	t.Run("Test nil select", func(t *testing.T) {
		q := NewQuery(&ctx).Model(model).Select(nil, nil)

		if q.Columns != nil {
			t.Fatal("Expected no column on select clause.")
		}
	})

	t.Run("Test select all columns", func(t *testing.T) {
		q := NewQuery(&ctx).Model(model).Select([]string{"*"}, nil)

		if q.Columns[0] != "*" {
			t.Fatal("Expected select *.")
		}
	})

	t.Run("Test select id and title", func(t *testing.T) {
		columns := []string{"id", "title"}

		q := NewQuery(&ctx).Model(model).Select(columns, nil)

		if len(q.Columns) != 2 {
			t.Fatal("Expected 2 columns on select clause.")
		}
	})

	t.Run("Test select unknown columns", func(t *testing.T) {
		columns := []string{"none", "unknown"}

		q := NewQuery(&ctx).Model(model).Select(columns, nil)

		if !q.HasError() {
			t.Fatalf("Expector error because \"%s\" and \"%s\" columns is not exists", "none", "unknown")
		}
	})
}
