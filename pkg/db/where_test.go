package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEq(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Eq("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=eq.1", q.GetUrl(), "the url should match")
}

func TestNotEq(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).NotEq("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=not.eq.1", q.GetUrl(), "the url should match")
}

func TestOrEq(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrEq("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&or=(id.eq.1)", q.GetUrl(), "the url should match")
}

func TestNeq(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Neq("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=neq.1", q.GetUrl(), "the url should match")
}

func TestNotNeq(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).NotNeq("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=not.neq.1", q.GetUrl(), "the url should match")
}

func TestOrNeq(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrNeq("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&or=(id.neq.1)", q.GetUrl(), "the url should match")
}

func TestLt(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Lt("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=lt.1", q.GetUrl(), "the url should match")
}

func TestNotLt(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).NotLt("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=not.lt.1", q.GetUrl(), "the url should match")
}

func TestOrLt(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrLt("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&or=(id.lt.1)", q.GetUrl(), "the url should match")
}

func TestLte(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Lte("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=lte.1", q.GetUrl(), "the url should match")
}

func TestNotLte(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).NotLte("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=not.lte.1", q.GetUrl(), "the url should match")
}

func TestOrLte(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrLte("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&or=(id.lte.1)", q.GetUrl(), "the url should match")
}

func TestGt(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Gt("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=gt.1", q.GetUrl(), "the url should match")
}

func TestNotGt(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).NotGt("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=not.gt.1", q.GetUrl(), "the url should match")
}

func TestOrGt(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrGt("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&or=(id.gt.1)", q.GetUrl(), "the url should match")
}

func TestGte(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Gte("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=gte.1", q.GetUrl(), "the url should match")
}

func TestNotGte(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).NotGte("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&id=not.gte.1", q.GetUrl(), "the url should match")
}

func TestOrGte(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrGte("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&or=(id.gte.1)", q.GetUrl(), "the url should match")
}

func TestIn(t *testing.T) {
	t.Run("where in int", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).In("popularity", []int{-5, 0, 7})

		if q.WhereAndList == nil {
			t.Error("Expected where clause not to be nil")
		}

		assert.Equalf(t, "/rest/v1/articles?select=*&popularity=in.(-5,0,7)", q.GetUrl(), "the url should match")
	})

	t.Run("where in uint", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).In("id", []uint{1, 2, 3})

		if q.WhereAndList == nil {
			t.Error("Expected where clause not to be nil")
		}

		assert.Equalf(t, "/rest/v1/articles?select=*&id=in.(1,2,3)", q.GetUrl(), "the url should match")
	})

	t.Run("where in float", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).In("price", []float64{0.25, 10.5, 7.75})

		if q.WhereAndList == nil {
			t.Error("Expected where clause not to be nil")
		}

		assert.Equalf(t, "/rest/v1/articles?select=*&price=in.(0.25,10.5,7.75)", q.GetUrl(), "the url should match")
	})

	t.Run("where in string", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).In("username", []string{"a", "b", "c"})

		if q.WhereAndList == nil {
			t.Error("Expected where clause not to be nil")
		}

		assert.Equalf(t, "/rest/v1/articles?select=*&username=in.(a,b,c)", q.GetUrl(), "the url should match")
	})

	t.Run("where in bool", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).In("is_allowed", []bool{true})

		if q.WhereAndList == nil {
			t.Error("Expected where clause not to be nil")
		}

		assert.Equalf(t, "/rest/v1/articles?select=*&is_allowed=in.(true)", q.GetUrl(), "the url should match")
	})

	t.Run("where not in", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).NotIn("id", []uint{1, 2, 3})

		if q.WhereAndList == nil {
			t.Error("Expected where clause not to be nil")
		}

		assert.Equalf(t, "/rest/v1/articles?select=*&id=not.in.(1,2,3)", q.GetUrl(), "the url should match")
	})
}

func TestOrIn(t *testing.T) {
	t.Run("where in int", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrIn("popularity", []int{-5, 0, 7})

		if q.WhereOrList == nil {
			t.Error("Expected where clause not to be nil")
		}

		assert.Equalf(t, "/rest/v1/articles?select=*&or=(popularity.in.(-5,0,7))", q.GetUrl(), "the url should match")
	})

	t.Run("where in uint", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrIn("id", []uint{1, 2, 3})

		if q.WhereOrList == nil {
			t.Error("Expected where clause not to be nil")
		}

		assert.Equalf(t, "/rest/v1/articles?select=*&or=(id.in.(1,2,3))", q.GetUrl(), "the url should match")
	})

	t.Run("where in float", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrIn("price", []float64{0.25, 10.5, 7.75})

		if q.WhereOrList == nil {
			t.Error("Expected where clause not to be nil")
		}

		assert.Equalf(t, "/rest/v1/articles?select=*&or=(price.in.(0.25,10.5,7.75))", q.GetUrl(), "the url should match")
	})

	t.Run("where in string", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrIn("username", []string{"a", "b", "c"})

		if q.WhereOrList == nil {
			t.Error("Expected where clause not to be nil")
		}

		assert.Equalf(t, "/rest/v1/articles?select=*&or=(username.in.(a,b,c))", q.GetUrl(), "the url should match")
	})

	t.Run("where in bool", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrIn("is_allowed", []bool{true})

		if q.WhereOrList == nil {
			t.Error("Expected where clause not to be nil")
		}

		assert.Equalf(t, "/rest/v1/articles?select=*&or=(is_allowed.in.(true))", q.GetUrl(), "the url should match")
	})
}

func TestLike(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Like("name", "%supa%")

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&name=like.*supa*", q.GetUrl(), "the url should match")
}

func TestNotLike(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).NotLike("name", "%supa%")

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&name=not.like.*supa*", q.GetUrl(), "the url should match")
}

func TestOrLike(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrLike("name", "%supa%")

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&or=(name.like.*supa*)", q.GetUrl(), "the url should match")
}

func TestIlike(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Ilike("name", "%supa%")

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&name=ilike.*supa*", q.GetUrl(), "the url should match")
}

func TestNotIlike(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).NotIlike("name", "%supa%")

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&name=not.ilike.*supa*", q.GetUrl(), "the url should match")
}

func TestOrIlike(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).OrIlike("name", "%supa%")

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&or=(name.ilike.*supa*)", q.GetUrl(), "the url should match")
}

func TestIs(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Is("is_featured", "true")

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*is_featured=is.true", q.GetUrl(), "the url should match")
}

func TestNotIs(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).NotIs("is_featured", "true")

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}

	assert.Equalf(t, "/rest/v1/articles?select=*&is_featured=is.true", q.GetUrl(), "the url should match")
}
