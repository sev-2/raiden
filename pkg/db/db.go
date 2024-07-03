package db

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/valyala/fasthttp"
)

type Query struct {
	Context     raiden.Context
	model       interface{}
	UseWhere    int
	Columns     []string
	EqList      *[]string
	NeqList     *[]string
	OrList      *[]string
	InList      *[]string
	LtList      *[]string
	LteList     *[]string
	GtList      *[]string
	GteList     *[]string
	LikeList    *[]string
	IlikeList   *[]string
	OrderList   *[]string
	LimitValue  int
	OffsetValue int
	Err         error
}

type ModelBase struct {
	raiden.ModelBase
}

type Where struct {
	column   string
	operator string
	value    any
}

func (q *Query) Error() error {
	return q.Err
}

func NewQuery(ctx raiden.Context) *Query {
	return &Query{
		Context: ctx,
	}
}

func (q *Query) Model(m interface{}) *Query {
	q.model = m
	return q
}

// Deprecated: Will be removed
func (m *ModelBase) NewQuery() *Query {
	return &Query{model: m}
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

func (q Query) Get() ([]byte, error) {

	url := q.GetUrl()

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "return=representation"

	resp, _, err := PostgrestRequest(q.Context, fasthttp.MethodGet, url, nil, headers)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (q Query) Single() ([]byte, error) {
	url := q.Limit(1).GetUrl()

	headers := make(map[string]string)

	headers["Accept"] = "application/vnd.pgrst.object+json"

	res, _, err := PostgrestRequest(q.Context, "GET", url, nil, headers)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (q Query) GetUrl() string {
	urlQuery := buildQueryURI(q)
	baseUrl := getConfig().SupabasePublicUrl
	url := fmt.Sprintf("%s/rest/v1/%s", baseUrl, urlQuery)
	return url
}

// Deprecated: Will be removed
func (m ModelBase) GetTable() string {
	t := reflect.TypeOf(m)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Name() == "" {
		return reflect.TypeOf(m).Field(0).Name
	}

	return t.Name()
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

	if q.EqList != nil && len(*q.EqList) > 0 {
		eqList := strings.Join(*q.EqList, "&")
		output += fmt.Sprintf("&%s", eqList)
	}

	if q.NeqList != nil && len(*q.NeqList) > 0 {
		neqList := strings.Join(*q.NeqList, "&")
		output += fmt.Sprintf("&%s", neqList)
	}

	if q.OrList != nil && len(*q.OrList) > 0 {
		orList := strings.Join(*q.OrList, "&")
		output += fmt.Sprintf("&%s", orList)
	}

	if q.InList != nil && len(*q.InList) > 0 {
		inList := strings.Join(*q.InList, "&")
		output += fmt.Sprintf("&%s", inList)
	}

	if q.LtList != nil && len(*q.LtList) > 0 {
		ltList := strings.Join(*q.LtList, "&")
		output += fmt.Sprintf("&%s", ltList)
	}

	if q.LteList != nil && len(*q.LteList) > 0 {
		ltList := strings.Join(*q.LteList, "&")
		output += fmt.Sprintf("&%s", ltList)
	}

	if q.GtList != nil && len(*q.GtList) > 0 {
		ltList := strings.Join(*q.GtList, "&")
		output += fmt.Sprintf("&%s", ltList)
	}

	if q.GteList != nil && len(*q.GteList) > 0 {
		ltList := strings.Join(*q.GteList, "&")
		output += fmt.Sprintf("&%s", ltList)
	}

	if q.LikeList != nil && len(*q.LikeList) > 0 {
		ltList := strings.Join(*q.LikeList, "&")
		output += fmt.Sprintf("&%s", ltList)
	}

	if q.IlikeList != nil && len(*q.IlikeList) > 0 {
		ltList := strings.Join(*q.IlikeList, "&")
		output += fmt.Sprintf("&%s", ltList)
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

func getConfig() *raiden.Config {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return nil
	}

	configFilePath := strings.Join([]string{currentDir, "app.yaml"}, string(os.PathSeparator))

	config, err := raiden.LoadConfig(&configFilePath)
	if err != nil {
		log.Println(err)
		return nil
	}

	return config
}
