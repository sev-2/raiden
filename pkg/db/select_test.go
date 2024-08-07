package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelect(t *testing.T) {
	t.Run("Test nil select", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Select(nil)

		if q.Columns != nil {
			t.Fatal("Expected no column on select clause.")
		}

		assert.Equal(t, q.GetUrl(), "/rest/v1/articles?select=*")
	})

	t.Run("Test select all columns", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Select([]string{"*"})

		if q.Columns[0] != "*" {
			t.Fatal("Expected select *.")
		}

		assert.Equal(t, q.GetUrl(), "/rest/v1/articles?select=*")
	})

	t.Run("Test select id and title", func(t *testing.T) {
		columns := []string{"id", "title"}

		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Select(columns)

		if len(q.Columns) != 2 {
			t.Fatal("Expected 2 columns on select clause.")
		}

		assert.Equal(t, q.GetUrl(), "/rest/v1/articles?select=id,title")
	})

	t.Run("Test select title with alias", func(t *testing.T) {
		columns := []string{"id", "name:title"}

		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Select(columns)

		if len(q.Columns) != 2 {
			t.Fatal("Expected 2 columns on select clause.")
		}

		assert.Equal(t, q.GetUrl(), "/rest/v1/articles?select=id,name:title")
	})

	t.Run("Test select unknown columns", func(t *testing.T) {
		columns := []string{"none", "unknown"}

		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Select(columns)

		if !q.HasError() {
			t.Fatalf("Expector error because \"%s\" and \"%s\" columns is not exists", "none", "unknown")
		}
	})
}
