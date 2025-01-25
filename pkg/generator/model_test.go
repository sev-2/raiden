package generator_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateModels(t *testing.T) {
	dir, err := os.MkdirTemp("", "model")
	assert.NoError(t, err)

	modelPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(modelPath)
	assert.NoError(t, err1)

	relationshipAction := objects.TablesRelationshipAction{
		UpdateAction:   "c",
		DeletionAction: "c",
	}

	tables := []*generator.GenerateModelInput{
		{
			Table: objects.Table{
				Name:   "test_table",
				Schema: "public",
				PrimaryKeys: []objects.PrimaryKey{
					{Name: "id"},
				},
				Columns: []objects.Column{
					{Name: "id", DataType: "integer", IsNullable: false},
					{Name: "name", DataType: "text", IsNullable: true},
				},
				RLSEnabled: true,
				RLSForced:  false,
			},
			Relations: []state.Relation{
				{
					RelationType: "one_to_many",
					Table:        "related_table",
					ForeignKey:   "test_table_id",
					PrimaryKey:   "id",
					Action:       &relationshipAction,
					Index:        &objects.Index{Schema: "public", Table: "related_table", Name: "test_table_id", Definition: "CREATE INDEX test_table_id ON related_table(test_table_id);"},
				},
			},
			Policies: objects.Policies{},
			ValidationTags: state.ModelValidationTag{
				"id":   "required",
				"name": "required",
			},
		},
	}

	err2 := generator.GenerateModels(dir, tables, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.FileExists(t, dir+"/internal/models/test_table.go")
}

func TestBuildRelationFields(t *testing.T) {
	table := objects.Table{
		Name: "profiles",
	}
	relationStr := `[{"Table":"places","Type":"[]*Places","RelationType":"manyToMany","PrimaryKey":"","ForeignKey":"","Tag":"","SourcePrimaryKey":"id","JoinsSourceForeignKey":"profile_id","TargetPrimaryKey":"id","JoinTargetForeignKey":"place_id","Through":"place_likes"},{"Table":"followers","Type":"[]*Followers","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"food_review_likes","Type":"[]*FoodReviewLikes","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"food_review_buddies","Type":"[]*FoodReviewBuddies","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"place_reviews","Type":"[]*PlaceReviews","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"place_likes","Type":"[]*PlaceLikes","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"collections","Type":"[]*Collections","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"chat_messages","Type":"[]*ChatMessages","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"place_review_likes","Type":"[]*PlaceReviewLikes","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"contact_numbers","Type":"[]*ContactNumbers","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"food_likes","Type":"[]*FoodLikes","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"place_review_buddies","Type":"[]*PlaceReviewBuddies","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"profile_badges","Type":"[]*ProfileBadges","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"place_check_ins","Type":"[]*PlaceCheckIns","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"food_reviews","Type":"[]*FoodReviews","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"posts","Type":"[]*Posts","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"activity_logs","Type":"[]*ActivityLogs","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"social_media_accounts","Type":"[]*SocialMediaAccounts","RelationType":"hasMany","PrimaryKey":"id","ForeignKey":"profile_id","Tag":""},{"Table":"places","Type":"[]*Places","RelationType":"manyToMany","PrimaryKey":"","ForeignKey":"","Tag":"","SourcePrimaryKey":"id","JoinsSourceForeignKey":"profile_id","TargetPrimaryKey":"id","JoinTargetForeignKey":"place_id","Through":"place_reviews"},{"Table":"place_reviews","Type":"[]*PlaceReviews","RelationType":"manyToMany","PrimaryKey":"","ForeignKey":"","Tag":"","SourcePrimaryKey":"id","JoinsSourceForeignKey":"profile_id","TargetPrimaryKey":"id","JoinTargetForeignKey":"place_review_id","Through":"place_review_likes"},{"Table":"collection_types","Type":"[]*CollectionTypes","RelationType":"manyToMany","PrimaryKey":"","ForeignKey":"","Tag":"","SourcePrimaryKey":"id","JoinsSourceForeignKey":"profile_id","TargetPrimaryKey":"id","JoinTargetForeignKey":"collection_type_id","Through":"collections"},{"Table":"foods","Type":"[]*Foods","RelationType":"manyToMany","PrimaryKey":"","ForeignKey":"","Tag":"","SourcePrimaryKey":"id","JoinsSourceForeignKey":"profile_id","TargetPrimaryKey":"id","JoinTargetForeignKey":"food_id","Through":"food_likes"},{"Table":"food_reviews","Type":"[]*FoodReviews","RelationType":"manyToMany","PrimaryKey":"","ForeignKey":"","Tag":"","SourcePrimaryKey":"id","JoinsSourceForeignKey":"profile_id","TargetPrimaryKey":"id","JoinTargetForeignKey":"food_review_id","Through":"food_review_likes"},{"Table":"places","Type":"[]*Places","RelationType":"manyToMany","PrimaryKey":"","ForeignKey":"","Tag":"","SourcePrimaryKey":"id","JoinsSourceForeignKey":"profile_id","TargetPrimaryKey":"id","JoinTargetForeignKey":"place_id","Through":"place_check_ins"},{"Table":"place_reviews","Type":"[]*PlaceReviews","RelationType":"manyToMany","PrimaryKey":"","ForeignKey":"","Tag":"","SourcePrimaryKey":"id","JoinsSourceForeignKey":"profile_id","TargetPrimaryKey":"id","JoinTargetForeignKey":"place_review_id","Through":"place_review_buddies"},{"Table":"food_reviews","Type":"[]*FoodReviews","RelationType":"manyToMany","PrimaryKey":"","ForeignKey":"","Tag":"","SourcePrimaryKey":"id","JoinsSourceForeignKey":"profile_id","TargetPrimaryKey":"id","JoinTargetForeignKey":"food_review_id","Through":"food_review_buddies"},{"Table":"badges","Type":"[]*Badges","RelationType":"manyToMany","PrimaryKey":"","ForeignKey":"","Tag":"","SourcePrimaryKey":"id","JoinsSourceForeignKey":"profile_id","TargetPrimaryKey":"id","JoinTargetForeignKey":"badge_id","Through":"profile_badges"},{"Table":"foods","Type":"[]*Foods","RelationType":"manyToMany","PrimaryKey":"","ForeignKey":"","Tag":"","SourcePrimaryKey":"id","JoinsSourceForeignKey":"profile_id","TargetPrimaryKey":"id","JoinTargetForeignKey":"food_id","Through":"food_reviews"}]`
	relations := make([]state.Relation, 0)
	err := json.Unmarshal([]byte(relationStr), &relations)
	assert.NoError(t, err)

	result := generator.BuildRelationFields(table, relations)

	mapTable := make(map[string]bool)
	for _, v := range result {
		_, exist := mapTable[v.Table]
		assert.False(t, exist)
		mapTable[v.Table] = true
	}
}
