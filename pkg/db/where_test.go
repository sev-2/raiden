package db

import (
	"testing"

	"github.com/sev-2/raiden"
)

func TestEq(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).Eq("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestOrEq(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).OrEq("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestNeq(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).Neq("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestOrNeq(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).OrNeq("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestLt(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).Lt("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestOrLt(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).OrLt("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestLte(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).Lte("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestOrLte(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).OrLte("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestGt(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).Gt("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestOrGt(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).OrGt("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestGte(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).Gte("id", 1)

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestOrGte(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).OrGte("id", 1)

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestIn(t *testing.T) {
	ctx := raiden.Ctx{}

	t.Run("where in int", func(t *testing.T) {
		q := NewQuery(&ctx).In("popularity", []int{-5, 0, 7})

		if q.WhereAndList == nil {
			t.Error("Expected where clause not to be nil")
		}
	})

	t.Run("where in uint", func(t *testing.T) {
		q := NewQuery(&ctx).In("id", []uint{1, 2, 3})

		if q.WhereAndList == nil {
			t.Error("Expected where clause not to be nil")
		}
	})

	t.Run("where in float", func(t *testing.T) {
		q := NewQuery(&ctx).In("price", []float64{0.25, 10.5, 7.75})

		if q.WhereAndList == nil {
			t.Error("Expected where clause not to be nil")
		}
	})

	t.Run("where in string", func(t *testing.T) {
		q := NewQuery(&ctx).In("username", []string{"a", "b", "c"})

		if q.WhereAndList == nil {
			t.Error("Expected where clause not to be nil")
		}
	})

	t.Run("where in bool", func(t *testing.T) {
		q := NewQuery(&ctx).In("is_allowed", []bool{true})

		if q.WhereAndList == nil {
			t.Error("Expected where clause not to be nil")
		}
	})
}

func TestOrIn(t *testing.T) {
	ctx := raiden.Ctx{}

	t.Run("where in int", func(t *testing.T) {
		q := NewQuery(&ctx).OrIn("popularity", []int{-5, 0, 7})

		if q.WhereOrList == nil {
			t.Error("Expected where clause not to be nil")
		}
	})

	t.Run("where in uint", func(t *testing.T) {
		q := NewQuery(&ctx).OrIn("id", []uint{1, 2, 3})

		if q.WhereOrList == nil {
			t.Error("Expected where clause not to be nil")
		}
	})

	t.Run("where in float", func(t *testing.T) {
		q := NewQuery(&ctx).OrIn("price", []float64{0.25, 10.5, 7.75})

		if q.WhereOrList == nil {
			t.Error("Expected where clause not to be nil")
		}
	})

	t.Run("where in string", func(t *testing.T) {
		q := NewQuery(&ctx).OrIn("username", []string{"a", "b", "c"})

		if q.WhereOrList == nil {
			t.Error("Expected where clause not to be nil")
		}
	})

	t.Run("where in bool", func(t *testing.T) {
		q := NewQuery(&ctx).OrIn("is_allowed", []bool{true})

		if q.WhereOrList == nil {
			t.Error("Expected where clause not to be nil")
		}
	})
}

func TestLike(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).Like("name", "%supa%")

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestOrLike(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).OrLike("name", "%supa%")

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestIlike(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).Ilike("name", "%supa%")

	if q.WhereAndList == nil {
		t.Error("Expected where clause not to be nil")
	}
}

func TestOrIlike(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).OrIlike("name", "%supa%")

	if q.WhereOrList == nil {
		t.Error("Expected where clause not to be nil")
	}
}
