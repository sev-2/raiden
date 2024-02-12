package raiden_test

import (
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
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

func (r *GetSubmissions) GetDefinition() string {
	return `BEGIN RETURN QUERY SELECT s.id, s.created_at, sc.name as sc_name, c.name as c_name FROM :s s INNER JOIN :sc sc ON s.scouter_id = sc.scouter_id INNER JOIN :c c ON s.candidate_id = c.candidate_id WHERE sc.name = :scouter_name AND c.name = :candidate_name; END;`
}

func TestCreateQuery(t *testing.T) {
	rpc := &GetSubmissions{}
	e := raiden.BuildRpc(rpc)
	assert.NoError(t, e)

	assert.Equal(t, "get_submissions", rpc.GetName())
	assert.Equal(t, "public", rpc.GetSchema())

	expectedCompleteQuery := "create or replace function public.get_submissions(scouter_name character varying, candidate_name text) returns table(id integer, created_at timestamp without time zone, sc_name character varying, c_name character varying) language plpgsql as $function$ begin return query select s.id, s.created_at, sc.name as sc_name, c.name as c_name from submission s inner join scouter sc on s.scouter_id = sc.scouter_id inner join candidate c on s.candidate_id = c.candidate_id where sc.name = scouter_name and c.name = candidate_name ; end; $function$"
	assert.Equal(t, expectedCompleteQuery, rpc.GetCompleteStmt())
}
