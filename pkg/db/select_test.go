package db

import (
	"testing"
)

func TestSelect(t *testing.T) {
	t.Run("Test nil select", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Select(nil, nil)

		if q.Columns != nil {
			t.Fatal("Expected no column on select clause.")
		}
	})

	t.Run("Test select all columns", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Select([]string{"*"}, nil)

		if q.Columns[0] != "*" {
			t.Fatal("Expected select *.")
		}
	})

	t.Run("Test select id and title", func(t *testing.T) {
		columns := []string{"id", "title"}

		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Select(columns, nil)

		if len(q.Columns) != 2 {
			t.Fatal("Expected 2 columns on select clause.")
		}
	})

	t.Run("Test select unknown columns", func(t *testing.T) {
		columns := []string{"none", "unknown"}

		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Select(columns, nil)

		if !q.HasError() {
			t.Fatalf("Expector error because \"%s\" and \"%s\" columns is not exists", "none", "unknown")
		}
	})
}
