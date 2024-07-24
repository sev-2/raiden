package db

import (
	"time"

	"github.com/sev-2/raiden"
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

var (
	articleMockModel = ArticleMockModel{}
)
