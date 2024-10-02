package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWith(t *testing.T) {
	articleMockModel := ArticleMockModel{}
	orderMockModel := OrdersMockModel{}

	t.Run("match url query for single relation", func(t *testing.T) {
		t.Run("without where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				Preload("User").
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,user:users!user_id(*)", url)
		})

		t.Run("with where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				Preload("User", "status", "eq", "approved").
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,user:users!user_id(*)&users.status=eq.approved", url)
		})
	})

	t.Run("match url query for multiple relations", func(t *testing.T) {
		t.Run("without where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(orderMockModel).
				Preload("UserBilling").
				Preload("UserAddress").
				GetUrl()

			assert.Equal(t, "/rest/v1/orders?select=*,user_billing:users!billing_id(*),user_address:users!address_id(*)", url)
		})

		t.Run("with where condition", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(orderMockModel).
				Preload("UserBilling", "status", "eq", "approved").
				Preload("UserAddress", "status", "eq", "approved").
				GetUrl()

			assert.Equal(t, "/rest/v1/orders?select=*,user_billing:users!billing_id(*),user_address:users!address_id(*)&users.status=eq.approved&users.status=eq.approved", url)
		})
	})
}
