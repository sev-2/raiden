package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	article := ArticleMockModel{
		Title:     "Foo title",
		Body:      "Foo body",
		CreatedAt: time.Now(),
	}

	_, err := NewQuery(&mockRaidenContext).
		Model(articleMockModel).
		Eq("id", 1).
		Update(article)

	assert.NoError(t, err)
}
