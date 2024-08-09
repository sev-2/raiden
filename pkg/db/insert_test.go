package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	article := ArticleMockModel{
		Id:        1,
		UserId:    1,
		Title:     "Foo title",
		Body:      "Foo body",
		CreatedAt: time.Now(),
	}

	_, err := NewQuery(&mockRaidenContext).
		Model(articleMockModel).
		Insert(article)

	assert.NoError(t, err)
}
