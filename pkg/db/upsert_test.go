package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpsert(t *testing.T) {

	articles := []ArticleMockModel{
		{
			Id:        1,
			UserId:    1,
			Title:     "Foo title",
			Body:      "Foo body",
			CreatedAt: time.Now(),
		},
		{
			Id:        2,
			UserId:    1,
			Title:     "Bar title",
			Body:      "Bar body",
			CreatedAt: time.Now(),
		},
	}

	// Convert to []interface{}
	var payload = make([]interface{}, len(articles))
	for i, v := range articles {
		payload[i] = v
	}

	opt := UpsertOptions{
		OnConflict: MergeDuplicates,
	}

	_, err := NewQuery(&mockRaidenContext).
		Model(articleMockModel).
		Upsert(payload, opt)

	assert.NoError(t, err)
}
