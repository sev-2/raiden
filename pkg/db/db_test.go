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

	Acl string `json:"-" read:"" write:""`

	User *UsersMockModel `json:"user,omitempty" join:"joinType:hasOne;primaryKey:id;foreignKey:user_id"`
}

type OrdersMockModel struct {
	raiden.ModelBase
	Id        int64     `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable:false;default:now()"`
	BillingId *int64    `json:"billing_id,omitempty" column:"name:billing_id;type:bigint;nullable"`
	AddressId *int64    `json:"address_id,omitempty" column:"name:address_id;type:bigint;nullable"`

	Metadata string `json:"-" schema:"public" tableName:"orders" rlsEnable:"true" rlsForced:"false"`

	Acl string `json:"-" read:"" write:""`

	// Relations
	UserBilling *UsersMockModel `json:"user_billing,omitempty" join:"joinType:hasOne;primaryKey:id;foreignKey:billing_id"`
	UserAddress *UsersMockModel `json:"user_address,omitempty" join:"joinType:hasOne;primaryKey:id;foreignKey:address_id"`
}

type UsersMockModel struct {
	ModelBase

	Id        int64     `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	Username  string    `json:"username,omitempty" column:"name:username;type:text;nullable:false"`
	TeamId    int64     `json:"team_id,omitempty" column:"name:team_id;type:bigint;nullable:false"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable:false;default:now()"`

	Metadata string `json:"-" schema:"public" tableName:"users" rlsEnable:"true" rlsForced:"false"`

	Acl string `json:"-" read:"" write:""`

	Team     *TeamsMockModel     `json:"team,omitempty" join:"joinType:hasOne;primaryKey:id;foreignKey:team_id"`
	Articles []*ArticleMockModel `json:"articles,omitempty" join:"joinType:hasMany;primaryKey:id;foreignKey:article_id"`
	Orders   []*OrdersMockModel  `json:"orders,omitempty" join:"joinType:manyToMany;through:orders;sourcePrimaryKey:id;sourceForeignKey:address_id;targetPrimaryKey:id;targetForeign:address_id"`
}

type TeamsMockModel struct {
	ModelBase

	Id             int64     `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	Name           string    `json:"name,omitempty" column:"name:name;type:text;nullable:false"`
	OrganizationId int64     `json:"organization_id,omitempty" column:"name:organization_id;type:bigint;nullable:false"`
	CreatedAt      time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable:false;default:now()"`

	Metadata string `json:"-" schema:"public" tableName:"teams" rlsEnable:"true" rlsForced:"false"`

	Acl string `json:"-" read:"" write:""`

	Organization *OrganizationsMockModel `json:"organization,omitempty" join:"joinType:hasOne;primaryKey:id;foreignKey:organization_id"`
	Users        []*UsersMockModel       `json:"users,omitempty" join:"joinType:hasMany;primaryKey:id;foreignKey:organization_id"`
}

type OrganizationsMockModel struct {
	ModelBase

	Id        int64     `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	Name      string    `json:"name,omitempty" column:"name:name;type:text;nullable:false"`
	CreatedAt time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable:false;default:now()"`

	Metadata string `json:"-" schema:"public" tableName:"organizations" rlsEnable:"true" rlsForced:"false"`

	Acl string `json:"-" read:"" write:""`

	Teams []*TeamsMockModel `json:"teams,omitempty" join:"joinType:hasMany;primaryKey:id;foreignKey:organization_id"`
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
	var articleMockModel = ArticleMockModel{}
	_, err := NewQuery(&mockRaidenContext).Model(articleMockModel).Single(&articleMockModel)

	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	var collection []ArticleMockModel
	_, err := NewQuery(&mockRaidenContext).Model(articleMockModel).Get(&collection)

	assert.NoError(t, err)
}
