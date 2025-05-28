package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWith(t *testing.T) {
	articleMockModel := ArticleMockModel{}
	orderMockModel := OrdersMockModel{}
	userMockModel := UsersMockModel{}

	t.Run("match url query for single relation", func(t *testing.T) {
		t.Run("many to one relation without where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				Preload("User").
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,user:users!user_id(*)", url)
		})

		t.Run("many to one relation with where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				Preload("User", "status", "eq", "approved").
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,user:users!user_id(*)&users.status=eq.approved", url)
		})

		t.Run("one to many relation without where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(userMockModel).
				Preload("Articles").
				GetUrl()

			assert.Equal(t, "/rest/v1/users?select=*,articles!article_id(*)", url)
		})

		t.Run("one to many relation with where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(userMockModel).
				Preload("Articles", "rating", "eq", "5").
				GetUrl()

			assert.Equal(t, "/rest/v1/users?select=*,articles!article_id(*)&articles.rating=eq.5", url)
		})
	})

	t.Run("match url query for multiple relations", func(t *testing.T) {
		t.Run("many to one relation and many to one relation without where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(orderMockModel).
				Preload("UserBilling").
				Preload("UserAddress").
				GetUrl()

			assert.Equal(t, "/rest/v1/orders?select=*,user_billing:users!billing_id(*),user_address:users!address_id(*)", url)
		})

		t.Run("many to one and many to one relation with where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(orderMockModel).
				Preload("UserBilling", "status", "eq", "approved").
				Preload("UserAddress", "status", "eq", "approved").
				GetUrl()

			assert.Equal(t, "/rest/v1/orders?select=*,user_billing:users!billing_id(*),user_address:users!address_id(*)&users.status=eq.approved&users.status=eq.approved", url)
		})

		t.Run("one to many relation and one to many relation without where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(userMockModel).
				Preload("Articles").
				Preload("Addresses").
				GetUrl()
			assert.Equal(t, "/rest/v1/users?select=*,articles!article_id(*),addresses!user_id(*)", url)
		})

		t.Run("one to many and one to many relation with where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(userMockModel).
				Preload("Articles", "rating", "eq", "5").
				Preload("Addresses", "address_id", "eq", "1").
				GetUrl()

			assert.Equal(t, "/rest/v1/users?select=*,articles!article_id(*),addresses!user_id(*)&articles.rating=eq.5&addresses.address_id=eq.1", url)
		})

		t.Run("one to many and many to one relation without where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(userMockModel).
				Preload("Articles").
				Preload("Team").
				GetUrl()

			assert.Equal(t, "/rest/v1/users?select=*,articles!article_id(*),team:teams!team_id(*)", url)
		})

		t.Run("one to many and many to one relation with where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(userMockModel).
				Preload("Articles", "rating", "eq", "5").
				Preload("Team", "name", "eq", "Engineering").
				GetUrl()

			assert.Equal(t, "/rest/v1/users?select=*,articles!article_id(*),team:teams!team_id(*)&articles.rating=eq.5&teams.name=eq.Engineering", url)
		})

	})

	t.Run("invalid relation", func(t *testing.T) {
		t.Run("invalid relation name", func(t *testing.T) {
			assert.Panics(t, func() {
				NewQuery(&mockRaidenContext).
					Model(articleMockModel).
					Preload("InvalidRelation")
			}, "Expected panic for invalid relation name")
		})

		t.Run("invalid relation when data type is not a struct", func(t *testing.T) {
			assert.Panics(t, func() {
				NewQuery(&mockRaidenContext).
					Model(articleMockModel).
					Preload("Title")
			}, "Expected panic for invalid relation with non-struct data type")
		})

		t.Run("invalid relation when data type is not a slice of struct", func(t *testing.T) {
			assert.Panics(t, func() {
				NewQuery(&mockRaidenContext).
					Model(articleMockModel).
					Preload("Tags")
			}, "Expected panic for invalid relation with non-slice of struct data type")
		})
	})
}
