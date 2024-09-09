package raiden_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type Scouter struct{}

type Candidate struct{}

type Submission struct{}

type GetSubmissionsParams struct {
	ScouterName   string `json:"scouter_name" column:"name:scouter_name;type:varchar"`
	CandidateName string `json:"candidate_name" column:"name:candidate_name;type:text"`
}
type GetSubmissionsItem struct {
	Id        int64     `json:"id" column:"name:id;type:integer"`
	CreatedAt time.Time `json:"created_at" column:"name:created_at;type:timestamp"`
	ScName    string    `json:"sc_name" column:"name:sc_name;type:varchar"`
	CName     string    `json:"c_name" column:"name:c_name;type:varchar"`
}

type GetSubmissionsResult []GetSubmissionsItem

type GetSubmissions struct {
	raiden.RpcBase
	Params *GetSubmissionsParams `json:"-"`
	Return GetSubmissionsResult  `json:"-"`
}

func (r *GetSubmissions) GetName() string {
	return "get_submissions"
}

func (r *GetSubmissions) UseParamPrefix() bool {
	return false
}

func (r *GetSubmissions) GetReturnType() raiden.RpcReturnDataType {
	return raiden.RpcReturnDataTypeTable
}

func (r *GetSubmissions) BindModels() {
	r.BindModel(Submission{}, "s").BindModel(Scouter{}, "sc").BindModel(Candidate{}, "c")
}

func (r *GetSubmissions) GetRawDefinition() string {
	return `BEGIN RETURN QUERY SELECT s.id, s.created_at, sc.name as sc_name, c.name as c_name FROM :s s INNER JOIN :sc sc ON s.scouter_id = sc.scouter_id INNER JOIN :c c ON s.candidate_id = c.candidate_id WHERE sc.name = :scouter_name AND c.name = :candidate_name; END;`
}

type RpcWithMissingReturn struct {
	raiden.RpcBase
	Params *GetSubmissionsParams `json:"-"`
}

func TestCreateQuery(t *testing.T) {
	rpc := &GetSubmissions{}
	e := raiden.BuildRpc(rpc)
	assert.NoError(t, e)

	assert.Equal(t, "get_submissions", rpc.GetName())
	assert.Equal(t, "public", rpc.GetSchema())

	expectedCompleteQuery := "create or replace function public.get_submissions(scouter_name character varying, candidate_name text) returns table(id integer, created_at timestamp without time zone, sc_name character varying, c_name character varying) language plpgsql set search_path = '' as $function$ begin return query select s.id, s.created_at, sc.name as sc_name, c.name as c_name from submission s inner join scouter sc on s.scouter_id = sc.scouter_id inner join candidate c on s.candidate_id = c.candidate_id where sc.name = scouter_name and c.name = candidate_name ; end; $function$"
	assert.Equal(t, expectedCompleteQuery, rpc.GetCompleteStmt())
}

func TestExecuteRpc(t *testing.T) {
	mockCtx := &mock.MockContext{
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{
				DeploymentTarget:    raiden.DeploymentTargetCloud,
				ProjectId:           "test-project-id",
				ProjectName:         "My Great Project",
				SupabaseApiBasePath: "/v1",
				SupabaseApiUrl:      "http://supabase.cloud.com",
				SupabasePublicUrl:   "http://supabase.cloud.com",
			}
		},
		RequestContextFn: func() *fasthttp.RequestCtx {
			rCtx := &fasthttp.RequestCtx{
				Request: fasthttp.Request{
					Header: fasthttp.RequestHeader{},
				},
			}

			rCtx.Request.Header.Set("Authorization", "Bearer some token")
			rCtx.Request.Header.Set("apiKey", "some api key")
			return rCtx
		},
	}

	mock := mock.MockSupabase{Cfg: mockCtx.Config()}
	mock.Activate()
	defer mock.Deactivate()

	err := mock.MockExecuteRpcWithExpectedResponse(200, "get_submissions", GetSubmissionsResult{})
	assert.NoError(t, err)

	rpc := &GetSubmissions{
		Params: &GetSubmissionsParams{
			ScouterName:   "test_1",
			CandidateName: "test_2",
		},
	}
	res, err := raiden.ExecuteRpc(mockCtx, rpc)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestExecuteRpcWithParams(t *testing.T) {
	requestCtx := &fasthttp.RequestCtx{
		Request: fasthttp.Request{},
	}
	requestCtx.Request.URI().QueryArgs().Set("limit", "10")

	mockCtx := &mock.MockContext{
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{
				DeploymentTarget:    raiden.DeploymentTargetCloud,
				ProjectId:           "test-project-id",
				ProjectName:         "My Great Project",
				SupabaseApiBasePath: "/v1",
				SupabaseApiUrl:      "http://supabase.cloud.com",
				SupabasePublicUrl:   "http://supabase.cloud.com",
			}
		},
		RequestContextFn: func() *fasthttp.RequestCtx {
			rCtx := &fasthttp.RequestCtx{
				Request: fasthttp.Request{
					Header: fasthttp.RequestHeader{},
				},
			}

			rCtx.Request.Header.Set("Authorization", "Bearer some token")
			rCtx.Request.Header.Set("apiKey", "some api key")
			return rCtx
		},
	}

	mock := mock.MockSupabase{Cfg: mockCtx.Config()}
	mock.Activate()
	defer mock.Deactivate()
	err := mock.MockExecuteRpcWithExpectedResponse(401, "get_submissions", map[string]interface{}{
		"message": "Invalid API key",
		"status":  401,
		"code":    "invalid_auth",
	})
	assert.NoError(t, err)

	rpc := &GetSubmissions{
		Params: &GetSubmissionsParams{
			ScouterName:   "test_1",
			CandidateName: "test_2",
		},
	}
	res, err := raiden.ExecuteRpc(mockCtx, rpc)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestExecuteRpcErrWithMissingReturn(t *testing.T) {
	mockCtx := &mock.MockContext{
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{
				DeploymentTarget:    raiden.DeploymentTargetCloud,
				ProjectId:           "test-project-id",
				ProjectName:         "My Great Project",
				SupabaseApiBasePath: "/v1",
				SupabaseApiUrl:      "http://supabase.cloud.com",
				SupabasePublicUrl:   "http://supabase.cloud.com",
			}
		},
		RequestContextFn: func() *fasthttp.RequestCtx {
			return &fasthttp.RequestCtx{}
		},
	}

	rpc := &RpcWithMissingReturn{}
	_, err := raiden.ExecuteRpc(mockCtx, rpc)

	expectedErr := &raiden.ErrorResponse{
		StatusCode: fasthttp.StatusInternalServerError,
		Details:    fmt.Sprintf("Struct %s doesn`t have Return field, define first because this attribute need for receive data from server", "RpcWithMissingReturn"),
		Message:    fmt.Sprintf("Undefined field Return in struct %s", "RpcWithMissingReturn"),
		Hint:       "Invalid Rpc",
		Code:       fasthttp.StatusMessage(fasthttp.StatusInternalServerError),
	}

	assert.Error(t, err)
	assert.EqualError(t, expectedErr, err.Error())

	mock := mock.MockSupabase{Cfg: mockCtx.Config()}
	mock.Activate()
	defer mock.Deactivate()
	err = mock.MockExecuteRpcWithExpectedResponse(200, "get_submissions", GetSubmissionsResult{})
	assert.NoError(t, err)
}

func TestRpcParamToGoType(t *testing.T) {
	tests := []struct {
		input    raiden.RpcParamDataType
		expected string
	}{
		{raiden.RpcParamDataTypeInteger, "int64"},
		{raiden.RpcParamDataTypeBigInt, "int64"},
		{raiden.RpcParamDataTypeReal, "float32"},
		{raiden.RpcParamDataTypeDoublePreci, "float64"},
		{raiden.RpcParamDataTypeText, "string"},
		{raiden.RpcParamDataTypeVarchar, "string"},
		{raiden.RpcParamDataTypeVarcharAlias, "string"},
		{raiden.RpcParamDataTypeBoolean, "bool"},
		{raiden.RpcParamDataTypeBytea, "[]byte"},
		{raiden.RpcParamDataTypeTimestamp, "time.Time"},
		{raiden.RpcParamDataTypeTimestampTZ, "time.Time"},
		{raiden.RpcParamDataTypeJSON, "map[string]interface{}"},
		{raiden.RpcParamDataTypeJSONB, "map[string]interface{}"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, raiden.RpcParamToGoType(tt.input))
	}
}

func TestGetValidRpcParamType(t *testing.T) {
	tests := []struct {
		input       string
		returnAlias bool
		expected    raiden.RpcParamDataType
		expectErr   bool
	}{
		{"integer", false, raiden.RpcParamDataTypeInteger, false},
		{"bigint", false, raiden.RpcParamDataTypeBigInt, false},
		{"real", false, raiden.RpcParamDataTypeReal, false},
		{"double precision", false, raiden.RpcParamDataTypeDoublePreci, false},
		{"varchar", false, raiden.RpcParamDataTypeVarchar, false},
		{"varchar", true, raiden.RpcParamDataTypeVarcharAlias, false},
		{"boolean", true, raiden.RpcParamDataTypeBoolean, false},
		{"bytea", true, raiden.RpcParamDataTypeBytea, false},
		{"timestamp", true, raiden.RpcParamDataTypeTimestampAlias, false},
		{"unsupported", false, "", true},
	}

	for _, tt := range tests {
		result, err := raiden.GetValidRpcParamType(tt.input, tt.returnAlias)
		if tt.expectErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestRpcReturnToGoType(t *testing.T) {
	tests := []struct {
		input    raiden.RpcReturnDataType
		expected string
	}{
		{raiden.RpcReturnDataTypeInteger, "int64"},
		{raiden.RpcReturnDataTypeBigInt, "int64"},
		{raiden.RpcReturnDataTypeReal, "float32"},
		{raiden.RpcReturnDataTypeDoublePreci, "float64"},
		{raiden.RpcReturnDataTypeText, "string"},
		{raiden.RpcReturnDataTypeVarchar, "string"},
		{raiden.RpcReturnDataTypeBoolean, "bool"},
		{raiden.RpcReturnDataTypeBytea, "[]byte"},
		{raiden.RpcReturnDataTypeTimestamp, "time.Time"},
		{raiden.RpcReturnDataTypeTimestampTZ, "time.Time"},
		{raiden.RpcReturnDataTypeJSON, "map[string]interface{}"},
		{raiden.RpcReturnDataTypeJSONB, "map[string]interface{}"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, raiden.RpcReturnToGoType(tt.input))
	}
}

func TestGetValidRpcReturnType(t *testing.T) {
	tests := []struct {
		input       string
		returnAlias bool
		expected    raiden.RpcReturnDataType
		expectErr   bool
	}{
		{"integer", false, raiden.RpcReturnDataTypeInteger, false},
		{"bigint", false, raiden.RpcReturnDataTypeBigInt, false},
		{"real", false, raiden.RpcReturnDataTypeReal, false},
		{"double precision", false, raiden.RpcReturnDataTypeDoublePreci, false},
		{"varchar", false, raiden.RpcReturnDataTypeVarchar, false},
		{"varchar", true, raiden.RpcReturnDataTypeVarcharAlias, false},
		{"boolean", true, raiden.RpcReturnDataTypeBoolean, false},
		{"bytea", true, raiden.RpcReturnDataTypeBytea, false},
		{"timestamp", true, raiden.RpcReturnDataTypeTimestampAlias, false},
		{"unsupported", false, "", true},
	}

	for _, tt := range tests {
		result, err := raiden.GetValidRpcReturnType(tt.input, tt.returnAlias)
		if tt.expectErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestGetValidRpcReturnNameDecl(t *testing.T) {
	tests := []struct {
		input       raiden.RpcReturnDataType
		returnAlias bool
		expected    string
		expectErr   bool
	}{
		{raiden.RpcReturnDataTypeInteger, false, "RpcReturnDataTypeInteger", false},
		{raiden.RpcReturnDataTypeVarchar, false, "RpcReturnDataTypeVarchar", false},
		{raiden.RpcReturnDataTypeVarchar, true, "RpcReturnDataTypeVarcharAlias", false},
		{raiden.RpcReturnDataTypeJSON, false, "RpcReturnDataTypeJSON", false},
		{raiden.RpcReturnDataTypeSetOf, false, "RpcReturnDataTypeSetOf", false},
		{raiden.RpcReturnDataTypeTable, false, "RpcReturnDataTypeTable", false},
		{raiden.RpcReturnDataTypeVoid, false, "RpcReturnDataTypeVoid", false},
		{raiden.RpcReturnDataType("unsupported"), false, "", true},
	}

	for _, tt := range tests {
		result, err := raiden.GetValidRpcReturnNameDecl(tt.input, tt.returnAlias)
		if tt.expectErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestMarshalRpcParamTag(t *testing.T) {
	tests := []struct {
		input    *raiden.RpcParamTag
		expected string
	}{
		{&raiden.RpcParamTag{Name: "id", Type: "integer", DefaultValue: "1"}, "name:id;type:integer;default:1"},
		{&raiden.RpcParamTag{Name: "name", Type: "varchar"}, "name:name;type:varchar"},
		{nil, ""},
	}

	for _, tt := range tests {
		result, err := raiden.MarshalRpcParamTag(tt.input)
		assert.NoError(t, err)
		assert.Equal(t, tt.expected, result)
	}
}

func TestUnmarshalRpcParamTag(t *testing.T) {
	tests := []struct {
		input     string
		expected  raiden.RpcParamTag
		expectErr bool
	}{
		{"name:id;type:integer;default:1", raiden.RpcParamTag{Name: "id", Type: "INTEGER", DefaultValue: "1"}, false},
		{"name:name;type:varchar", raiden.RpcParamTag{Name: "name", Type: "VARCHAR"}, false},
		{"type:rand;", raiden.RpcParamTag{}, true},
	}

	for _, tt := range tests {
		result, err := raiden.UnmarshalRpcParamTag(tt.input)
		if tt.expectErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}
