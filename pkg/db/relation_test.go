package db

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func TestWith(t *testing.T) {
	resource.RegisterModels(
		ArticleMockModel{},
		UsersMockModel{},
		TeamsMockModel{},
		OrganizationsMockModel{},
	)

	t.Run("match url query for single relation", func(t *testing.T) {
		t.Run("without selected columns", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				With("UsersMockModel", nil).
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,users(*)", url)
		})

		t.Run("with selected columns", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				With(
					"UsersMockModel",
					map[string][]string{
						"users": {"id", "username"},
					},
				).
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,users(id,username)", url)
		})

		t.Run("with selected columns and aliases", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				With(
					"UsersMockModel",
					map[string][]string{
						"users": {"id", "userid:username"},
					},
				).
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,users(id,userid:username)", url)
		})
	})

	t.Run("match url query for two-nested relation", func(t *testing.T) {
		t.Run("without selected columns", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				With("UsersMockModel.TeamsMockModel", nil).
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,users(*,teams(*))", url)
		})

		t.Run("with selected columns", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				With(
					"UsersMockModel.TeamsMockModel",
					map[string][]string{
						"users": {"id", "username"},
						"teams": {"id", "name"},
					},
				).
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,users(id,username,teams(id,name))", url)
		})

		t.Run("with selected columns and aliases", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				With(
					"UsersMockModel.TeamsMockModel",
					map[string][]string{
						"users": {"id", "userid:username"},
						"teams": {"id", "team_name:name"},
					},
				).
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,users(id,userid:username,teams(id,team_name:name))", url)
		})
	})

	t.Run("match url query for three-nested relation", func(t *testing.T) {
		t.Run("without selected columns", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				With("UsersMockModel.TeamsMockModel.OrganizationsMockModel", nil).
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,users(*,teams(*,organizations(*)))", url)
		})

		t.Run("with selected columns", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				With(
					"UsersMockModel.TeamsMockModel.OrganizationsMockModel",
					map[string][]string{
						"users":         {"id", "username"},
						"teams":         {"id", "name"},
						"organizations": {"id", "name"},
					},
				).
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,users(id,username,teams(id,name,organizations(id,name)))", url)
		})

		t.Run("with selected columns and aliases", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(articleMockModel).
				With(
					"UsersMockModel.TeamsMockModel.OrganizationsMockModel",
					map[string][]string{
						"users":         {"id", "userid:username"},
						"teams":         {"id", "team_name:name"},
						"organizations": {"id", "org_name:name"},
					},
				).
				GetUrl()

			assert.Equal(t, "/rest/v1/articles?select=*,users(id,userid:username,teams(id,team_name:name,organizations(id,org_name:name)))", url)
		})
	})

	t.Run("match url query with foreign key", func(t *testing.T) {

		t.Run("without selected column", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(OrdersMockModel{}).
				With(
					"UsersMockModel",
					map[string][]string{
						"users!address_id": {},
					},
				).
				GetUrl()

			assert.Equal(t, "/rest/v1/orders?select=*,users!address_id(*)", url)
		})

		t.Run("with all columns", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(OrdersMockModel{}).
				With(
					"UsersMockModel",
					map[string][]string{
						"users!address_id": {"*"},
					},
				).
				GetUrl()

			assert.Equal(t, "/rest/v1/orders?select=*,users!address_id(*)", url)
		})

		t.Run("with selected id", func(t *testing.T) {
			url := NewQuery(&mockRaidenContext).
				Model(OrdersMockModel{}).
				With(
					"UsersMockModel",
					map[string][]string{
						"users!address_id": {"id", "username"},
					},
				).
				GetUrl()

			assert.Equal(t, "/rest/v1/orders?select=*,users!address_id(id,username)", url)
		})
	})
}
