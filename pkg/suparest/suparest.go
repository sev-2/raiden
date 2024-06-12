package suparest

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/valyala/fasthttp"
)

type Query struct {
	model       *ModelBase
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

func (m *ModelBase) NewQuery() *Query {
	return &Query{model: m}
}

func (m *ModelBase) Execute() (model *ModelBase) {
	return m
}

func (q Query) Get() ([]byte, error) {

	url := q.GetUrl()

	client := &fasthttp.Client{}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)

	serviceKey := getConfig().ServiceKey
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		log.Fatalf("Error making GET request: %s\n", err)
	}

	body := resp.Body()

	return body, nil
}

func (q Query) Single() ([]byte, error) {
	result, err := q.Limit(1).Get()
	if err != nil {
		return nil, err
	}

	var object []any
	err = json.Unmarshal(result, &object)
	if err != nil {
		return nil, err
	}

	if len(object) > 0 {
		result, err := json.Marshal(object[0])
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	return result, nil
}

func (q Query) GetUrl() string {
	urlQuery := buildQueryURI(q)
	baseUrl := getConfig().SupabasePublicUrl
	url := fmt.Sprintf("%s/rest/v1/%s", baseUrl, urlQuery)
	return url
}

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

	output = fmt.Sprintf("%s?", q.model.GetTable())

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