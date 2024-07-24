package db

import (
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

var mockRaidenContext = raiden.Ctx{
	RequestCtx: &fasthttp.RequestCtx{},
}

type ArticleMockModel struct {
	ModelBase

	Id        int64     `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	UserId    int64     `json:"user_id,omitempty" column:"name:user_id;type:bigint;nullable:false"`
	Title     string    `json:"title,omitempty" column:"name:title;type:text;nullable:false"`
	Body      string    `json:"body,omitempty" column:"name:body;type:text;nullable:true"`
	Rating    int64     `json:"rating,omitempty" column:"name:rating;type:bigint;nullable:false"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable:false;default:now()"`

	Metadata string `json:"-" schema:"public" tableName:"articles" rlsEnable:"true" rlsForced:"false"`
}

type CommentsMockModel struct {
	ModelBase

	Id        int64     `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	UserId    int64     `json:"user_id,omitempty" column:"name:user_id;type:bigint;nullable:false"`
	ArticleId int64     `json:"article_id,omitempty" column:"name:article_id;type:bigint;nullable:false"`
	Body      string    `json:"body,omitempty" column:"name:body;type:text;nullable:true"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable:false;default:now()"`

	Metadata string `json:"-" schema:"public" tableName:"comments" rlsEnable:"true" rlsForced:"false"`
}

type UsersMockModel struct {
	ModelBase

	Id        int64     `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	Username  string    `json:"username,omitempty" column:"name:username;type:text;nullable:false"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable:false;default:now()"`

	Metadata string `json:"-" schema:"public" tableName:"comments" rlsEnable:"true" rlsForced:"false"`
}

var (
	articleMockModel = ArticleMockModel{}
)

func TestGetTable(t *testing.T) {
	t.Run("valid table name", func(t *testing.T) {
		table := GetTable(articleMockModel)
		assert.Equal(t, "articles", table)
	})

	t.Run("invalid table name", func(t *testing.T) {
		invalidModel := struct{}{}

		table := GetTable(invalidModel)
		assert.Equal(t, "", table)
	})
}

func TestSingle(t *testing.T) {
	_, err := NewQuery(&mockRaidenContext).Model(articleMockModel).Single()

	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	_, err := NewQuery(&mockRaidenContext).Model(articleMockModel).Get()

	assert.NoError(t, err)
}
