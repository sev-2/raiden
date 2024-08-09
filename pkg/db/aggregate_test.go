package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	count, _ := NewQuery(&mockRaidenContext).Model(articleMockModel).Count()

	assert.IsType(t, 0, count)
}

func TestCountWithOptionsExact(t *testing.T) {
	count, _ := NewQuery(&mockRaidenContext).Model(articleMockModel).Count(CountOptions{Count: "exact"})

	assert.IsType(t, 0, count)
}

func TestCountWithOptionsPlanned(t *testing.T) {
	count, _ := NewQuery(&mockRaidenContext).Model(articleMockModel).Count(CountOptions{Count: "planned"})

	assert.IsType(t, 0, count)
}

func TestCountWithOptionsEstimated(t *testing.T) {
	count, _ := NewQuery(&mockRaidenContext).Model(articleMockModel).Count(CountOptions{Count: "estimated"})

	assert.IsType(t, 0, count)
}

func TestCountWithWrongOptions(t *testing.T) {
	count, _ := NewQuery(&mockRaidenContext).Model(articleMockModel).Count(CountOptions{Count: "wrong"})

	assert.IsType(t, 0, count)
}

func TestSum(t *testing.T) {
	t.Run("sum", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Sum("rating", "")
		assert.Equalf(t, "/rest/v1/articles?select=rating.sum()", q.GetUrl(), "the url should match")
	})

	t.Run("sum with alias", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Sum("rating", "rate")
		assert.Equalf(t, "/rest/v1/articles?select=rate:rating.sum()", q.GetUrl(), "the url should match")
	})
}

func TestAvg(t *testing.T) {
	t.Run("avg", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Avg("rating", "")
		assert.Equalf(t, "/rest/v1/articles?select=rating.avg()", q.GetUrl(), "the url should match")
	})

	t.Run("avg with alias", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Avg("rating", "rate")
		assert.Equalf(t, "/rest/v1/articles?select=rate:rating.avg()", q.GetUrl(), "the url should match")
	})
}

func TestMin(t *testing.T) {
	t.Run("min", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Min("rating", "")
		assert.Equalf(t, "/rest/v1/articles?select=rating.min()", q.GetUrl(), "the url should match")
	})

	t.Run("min with alias", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Min("rating", "rate")
		assert.Equalf(t, "/rest/v1/articles?select=rate:rating.min()", q.GetUrl(), "the url should match")
	})
}

func TestMax(t *testing.T) {
	t.Run("max", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Max("rating", "")
		assert.Equalf(t, "/rest/v1/articles?select=rating.max()", q.GetUrl(), "the url should match")
	})

	t.Run("max with alias", func(t *testing.T) {
		q := NewQuery(&mockRaidenContext).Model(articleMockModel).Max("rating", "rate")
		assert.Equalf(t, "/rest/v1/articles?select=rate:rating.max()", q.GetUrl(), "the url should match")
	})
}
