package state_test

import (
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

type Scouter struct{}

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

func TestExtractRpc(t *testing.T) {
	rpcStates := []state.RpcState{
		{Function: objects.Function{Name: "existing_rpc"}},
	}

	rpc1 := &GetSubmissions{}
	e := raiden.BuildRpc(rpc1)
	assert.NoError(t, e)

	appRpcs := []raiden.Rpc{
		rpc1,
	}

	result, err := state.ExtractRpc(rpcStates, appRpcs)
	assert.NoError(t, err)
	assert.Len(t, result.Existing, 0)
	assert.Len(t, result.New, 1)
	assert.Len(t, result.Delete, 1)

	// Test for deletion
	appRpcs = []raiden.Rpc{}
	result, err = state.ExtractRpc(rpcStates, appRpcs)
	assert.NoError(t, err)
	assert.Len(t, result.Existing, 0)
	assert.Len(t, result.New, 0)
	assert.Len(t, result.Delete, 1)
}

func TestBindRpcFunction(t *testing.T) {
	rpc := &GetSubmissions{}
	e := raiden.BuildRpc(rpc)
	assert.NoError(t, e)
	fn := objects.Function{}

	err := state.BindRpcFunction(rpc, &fn)
	assert.NoError(t, err)
	assert.Equal(t, "get_submissions", fn.Name)
	assert.Equal(t, "public", fn.Schema)
	assert.Equal(t, "create or replace function public.get_submissions(scouter_name character varying, candidate_name text) returns table(id integer, created_at timestamp without time zone, sc_name character varying, c_name character varying) language plpgsql as $function$ begin return query select s.id, s.created_at, sc.name as sc_name, c.name as c_name from submission s inner join scouter sc on s.scouter_id = sc.scouter_id inner join candidate c on s.candidate_id = c.candidate_id where sc.name = scouter_name and c.name = candidate_name ; end; $function$", fn.CompleteStatement)
}

func TestExtractRpcResult_ToDeleteFlatMap(t *testing.T) {
	extractRpcResult := state.ExtractRpcResult{
		Delete: []objects.Function{
			{Name: "rpc1"},
			{Name: "rpc2"},
		},
	}

	mapData := extractRpcResult.ToDeleteFlatMap()
	assert.Len(t, mapData, 2)
	assert.Contains(t, mapData, "rpc1")
	assert.Contains(t, mapData, "rpc2")
	assert.Equal(t, "rpc1", mapData["rpc1"].Name)
	assert.Equal(t, "rpc2", mapData["rpc2"].Name)
}
