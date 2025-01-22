package db

import (
	"flag"
	"fmt"
	"reflect"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/valyala/fasthttp"
)

type Query struct {
	Context      raiden.Context
	model        interface{}
	Columns      []string
	Relations    []string
	WhereAndList *[]string
	WhereOrList  *[]string
	IsList       *[]string
	OrderList    *[]string
	LimitValue   int
	OffsetValue  int
	Errors       []error
	ByPass       bool
}

type ModelBase struct {
	raiden.ModelBase
}

func (q *Query) HasError() bool {
	return len(q.Errors) > 0
}

func NewQuery(ctx raiden.Context) *Query {
	return &Query{
		Context: ctx,
		ByPass:  false,
	}
}

// Model is the From alias
func (q *Query) Model(m interface{}) *Query {
	return q.From(m)
}

func (q *Query) From(m interface{}) *Query {
	q.model = m
	return q
}

func (q *Query) AsSystem() *Query {
	q.ByPass = true
	return q
}

func GetTable(m interface{}) string {
	t := reflect.TypeOf(m)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field, found := t.FieldByName("Metadata")
	if !found {
		fmt.Println("Field \"tableName\" is not found")
		return ""
	}

	tableName := field.Tag.Get("tableName")

	return tableName
}

func (m *ModelBase) Execute() (model *ModelBase) {
	return m
}

func (q Query) Get(collection interface{}) error {

	url := q.GetUrl()

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "return=representation"

	_, err := PostgrestRequestBind(q.Context, fasthttp.MethodGet, url, nil, headers, q.ByPass, collection)
	if err != nil {
		return err
	}

	return nil
}

func (q Query) Single(model interface{}) error {
	url := q.Limit(1).GetUrl()

	headers := make(map[string]string)

	headers["Accept"] = "application/vnd.pgrst.object+json"

	_, err := PostgrestRequestBind(q.Context, "GET", url, nil, headers, q.ByPass, model)
	if err != nil {
		return err
	}

	return nil
}

func (q Query) GetUrl() string {
	urlQuery := buildQueryURI(q)

	var baseUrl string
	if flag.Lookup("test.v") != nil {
		baseUrl = ""
	} else {
		baseUrl = getConfig().SupabasePublicUrl

		if getConfig().Mode == raiden.SvcMode {
			baseUrl = getConfig().PostgRestUrl
		}
	}

	url := fmt.Sprintf("%s/rest/v1/%s", baseUrl, urlQuery)

	if getConfig().Mode == raiden.SvcMode {
		url = fmt.Sprintf("%s/%s", baseUrl, urlQuery)
	}

	return url
}

func buildQueryURI(q Query) string {
	var output string

	output = fmt.Sprintf("%s?", GetTable(q.model))

	if len(q.Columns) > 0 {
		columns := strings.Join(q.Columns, ",")
		output += fmt.Sprintf("select=%s", columns)
	} else {
		output += "select=*"
	}

	if len(q.Relations) > 0 {
		output += "," + strings.Join(q.Relations, ",")
	}

	if q.WhereAndList != nil && len(*q.WhereAndList) > 0 {
		eqList := strings.Join(*q.WhereAndList, "&")
		output += fmt.Sprintf("&%s", eqList)
	}

	if q.WhereOrList != nil && len(*q.WhereOrList) > 0 {
		list := strings.Join(*q.WhereOrList, ",")
		output += fmt.Sprintf("&or=(%s)", list)
	}

	if q.IsList != nil && len(*q.IsList) > 0 {
		list := strings.Join(*q.IsList, ",")
		output += fmt.Sprintf("&%s", list)
	}

	if q.OrderList != nil && len(*q.OrderList) > 0 {
		orders := strings.Join(*q.OrderList, ",")
		output += fmt.Sprintf("&order=%s", orders)
	}

	if q.LimitValue > 0 {
		output += fmt.Sprintf("&limit=%v", q.LimitValue)
	}

	if q.OffsetValue > 0 {
		output += fmt.Sprintf("&offset=%v", q.OffsetValue)
	}

	return output
}
