package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRpcExtractParam(t *testing.T) {
	fn := objects.Function{
		Args: []objects.FunctionArg{
			{
				Mode:       "in",
				Name:       "in_candidate_name",
				TypeId:     1043,
				HasDefault: true,
			},
			{
				Mode:       "in",
				Name:       "in_voter_name",
				TypeId:     1043,
				HasDefault: true,
			},
		},
		ArgumentTypes: "in_candidate_name character varying DEFAULT 'anon'::character varying, in_voter_name character varying DEFAULT 'anon'::character varying",
	}

	param, usePrefix, err := generator.ExtractRpcParam(&fn)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(param))
	assert.Equal(t, true, usePrefix)

	assert.Equal(t, "candidate_name", param[0].Name)
	assert.Equal(t, raiden.RpcParamDataTypeVarcharAlias, param[0].Type)

	assert.Equal(t, "voter_name", param[1].Name)
	assert.Equal(t, raiden.RpcParamDataTypeVarcharAlias, param[1].Type)
}

func TestExtractRpcData(t *testing.T) {
	fn := objects.Function{
		Schema:   "public",
		Name:     "get_submissions",
		Language: "plpgsql",
		Definition: `
			BEGIN
				RETURN QUERY
					SELECT s.submission_id, s.submission_date, sc.scouter_name, c.candidate_name
					FROM submission s
					INNER JOIN scouter sc ON s.scouter_id = sc.scouter_id
					INNER JOIN candidate c ON s.candidate_id = c.candidate_id
					WHERE sc.scouter_name = scouter_name
					AND c.candidate_name = candidate_name;
			END;
		`,
		CompleteStatement: `
		CREATE OR REPLACE FUNCTION get_submissions(scouter_name VARCHAR, candidate_name VARCHAR)
		RETURNS TABLE (
			submission_id INT,
			submission_date TIMESTAMP,
			scouter_name VARCHAR,
			candidate_name VARCHAR
		) AS $$
		BEGIN
			RETURN QUERY
				SELECT s.submission_id, s.submission_date, sc.scouter_name, c.candidate_name
				FROM submission s
				INNER JOIN scouter sc ON s.scouter_id = sc.scouter_id
				INNER JOIN candidate c ON s.candidate_id = c.candidate_id
				WHERE sc.scouter_name = scouter_name
				AND c.candidate_name = candidate_name;
		END;
		$$ LANGUAGE plpgsql;
		`,
		ReturnType:      "TABLE(id integer, created_at timestamp without time zone, sc_name character varying, c_name character varying)",
		Behavior:        "VOLATILE",
		SecurityDefiner: false,
	}

	result, err := generator.ExtractRpcFunction(&fn)
	assert.NoError(t, err)

	assert.Equal(t, fn.Name, result.Rpc.Name)
	assert.Equal(t, raiden.DefaultRpcSchema, result.Rpc.Schema)
	assert.Equal(t, raiden.RpcBehaviorVolatile, result.Rpc.Behavior)

	assert.Equal(t, raiden.RpcSecurityTypeInvoker, result.Rpc.SecurityType)
	assert.Equal(t, raiden.RpcReturnDataTypeTable, result.Rpc.ReturnType)
	assert.Equal(t, fn.ReturnType, result.OriginalReturnType)
	assert.Equal(t, 3, len(result.MapScannedTable))
}

func TestExtractRpcTable(t *testing.T) {
	definition := `
		BEGIN
			RETURN QUERY 
				SELECT s.submission_id, s.submission_date, sc.scouter_name, c.candidate_name
				FROM submission s
				INNER JOIN scouter sc ON s.scouter_id = sc.scouter_id
				INNER JOIN candidate c ON s.candidate_id = c.candidate_id
				WHERE sc.scouter_name = scouter_name
				AND c.candidate_name = candidate_name;
		END;
	`

	_, mapTable, err := generator.ExtractRpcTable(definition)
	assert.NoError(t, err)

	submission, isSubmissionExist := mapTable["submission"]
	assert.True(t, isSubmissionExist)
	assert.Equal(t, "submission", submission.Name)
	assert.Equal(t, "s", submission.Alias)
	assert.Equal(t, 2, len(submission.Relation))

	candidate, isCandidateExist := mapTable["candidate"]
	assert.True(t, isCandidateExist)
	assert.Equal(t, "candidate", candidate.Name)
	assert.Equal(t, "c", candidate.Alias)
	assert.Equal(t, 1, len(candidate.Relation))

	scouter, isScouterExist := mapTable["scouter"]
	assert.True(t, isScouterExist)
	assert.Equal(t, "scouter", scouter.Name)
	assert.Equal(t, "sc", scouter.Alias)
	assert.Equal(t, 1, len(scouter.Relation))
}

func TestExtractRpcSingleTable(t *testing.T) {
	definition := `
	begin
	    return query select  * from todo;
	end;`

	_, mapTable, err := generator.ExtractRpcTable(definition)
	assert.NoError(t, err)

	table, isTableExist := mapTable["todo"]
	assert.True(t, isTableExist)
	assert.NotNil(t, table)
	assert.Equal(t, "todo", table.Name)
}

func TestNormalizeTableAlias(t *testing.T) {
	mapAlias := map[string]*generator.RpcScannedTable{
		"submission": {
			Name:  "submission",
			Alias: "s",
		},
		"scouter": {
			Name:  "scouter",
			Alias: "",
		},
	}

	err := generator.RpcNormalizeTableAliases(mapAlias)
	assert.NoError(t, err)
	assert.Equal(t, "sc", mapAlias["scouter"].Alias)
}

func TestExtractRpcWithPrefix(t *testing.T) {
	fn := objects.Function{
		Schema:   "public",
		Name:     "get_submissions",
		Language: "plpgsql",
		Definition: `begin return query 
		select s.id, s.created_at, sc.name as sc_name, c.name as c_name 
		from submission s
		inner join scouter sc on s.scouter_id = sc.scouter_id 
		inner join candidate c on s.candidate_id = c.candidate_id 
		where sc.name = in_scouter_name and c.name = in_candidate_name ; end;
		`,
		CompleteStatement: `
		CREATE OR REPLACE FUNCTION public.get_submissions(in_scouter_name character varying, in_candidate_name character varying)\n 
		RETURNS TABLE(id integer, created_at timestamp without time zone, sc_name character varying, c_name character varying)\n 
		LANGUAGE plpgsql\n
		AS $function$ 
			begin return query 
				select s.id, s.created_at, sc.name as sc_name, c.name as c_name from submission s 
				inner join scouter sc on s.scouter_id = sc.scouter_id 
				inner join candidate c on s.candidate_id = c.candidate_id 
				where sc.name = in_scouter_name and c.name = in_candidate_name ; end; 
		$function$\n
		`,
		Args: []objects.FunctionArg{
			{
				Mode:       "in",
				Name:       "in_scouter_name",
				TypeId:     1043,
				HasDefault: false,
			},
			{
				Mode:       "in",
				Name:       "in_candidate_name",
				TypeId:     1043,
				HasDefault: false,
			},
			{
				Mode:       "table",
				Name:       "id",
				TypeId:     23,
				HasDefault: false,
			},
			{
				Mode:       "table",
				Name:       "created_at",
				TypeId:     23,
				HasDefault: false,
			},
			{
				Mode:       "table",
				Name:       "sc_name",
				TypeId:     23,
				HasDefault: false,
			},
			{
				Mode:       "table",
				Name:       "c_name",
				TypeId:     23,
				HasDefault: false,
			},
		},
		ArgumentTypes:          "in_scouter_name character varying, in_candidate_name character varying",
		IdentityArgumentTypes:  "in_scouter_name character varying, in_candidate_name character varying",
		ReturnTypeID:           2249,
		ReturnType:             "TABLE(id integer, created_at timestamp without time zone, sc_name character varying, c_name character varying)",
		ReturnTypeRelationID:   0,
		IsSetReturningFunction: true,
		Behavior:               string(raiden.RpcBehaviorVolatile),
		SecurityDefiner:        false,
		ConfigParams:           nil,
	}

	result, err := generator.ExtractRpcFunction(&fn)
	assert.NoError(t, err)

	assert.Equal(t, fn.Name, result.Rpc.Name)
	assert.Equal(t, raiden.DefaultRpcSchema, result.Rpc.Schema)
	assert.Equal(t, raiden.RpcBehaviorVolatile, result.Rpc.Behavior)

	assert.Equal(t, raiden.RpcSecurityTypeInvoker, result.Rpc.SecurityType)
	assert.Equal(t, raiden.RpcReturnDataTypeTable, result.Rpc.ReturnType)
	assert.Equal(t, fn.ReturnType, result.OriginalReturnType)
	assert.Equal(t, 3, len(result.MapScannedTable))

	expectedDefinition := "begin return query select s.id, s.created_at, sc.name as sc_name, c.name as c_name from :s s inner join :sc sc on s.scouter_id = sc.scouter_id inner join :c c on s.candidate_id = c.candidate_id where sc.name = :scouter_name and c.name = :candidate_name ; end;"
	assert.Equal(t, expectedDefinition, result.Rpc.Definition)
}

func TestExtractQueryWithWrite(t *testing.T) {
	definition := `
	BEGIN
		drop table if exists _temptbl ; create temp table _temptbl ( id int, name varchar(250), description text, address varchar(200), latitude double precision, longitude double precision, price varchar(200), open_hours json, images json, google_maps_id int, created_at timestamp, updated_at timestamp, review_count_diff int, check_in_count_diff int, distance double precision ) on commit drop ; 
	END;
`

	_, mapTable, err := generator.ExtractRpcTable(definition)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(mapTable))
}

func TestGenerateRpc(t *testing.T) {
	fns := []objects.Function{
		{
			Schema:   "public",
			Name:     "get_submissions",
			Language: "plpgsql",
			Definition: `begin return query 
			select s.id, s.created_at, sc.name as sc_name, c.name as c_name 
			from submission s
			inner join scouter sc on s.scouter_id = sc.scouter_id 
			inner join candidate c on s.candidate_id = c.candidate_id 
			where sc.name = in_scouter_name and c.name = in_candidate_name ; end;
			`,
			CompleteStatement: `
			CREATE OR REPLACE FUNCTION public.get_submissions(in_scouter_name character varying, in_candidate_name character varying)\n 
			RETURNS TABLE(id integer, created_at timestamp without time zone, sc_name character varying, c_name character varying)\n 
			LANGUAGE plpgsql\n
			AS $function$ 
				begin return query 
					select s.id, s.created_at, sc.name as sc_name, c.name as c_name from submission s 
					inner join scouter sc on s.scouter_id = sc.scouter_id 
					inner join candidate c on s.candidate_id = c.candidate_id 
					where sc.name = in_scouter_name and c.name = in_candidate_name ; end; 
			$function$\n
			`,
			Args: []objects.FunctionArg{
				{
					Mode:       "in",
					Name:       "in_scouter_name",
					TypeId:     1043,
					HasDefault: false,
				},
				{
					Mode:       "in",
					Name:       "in_candidate_name",
					TypeId:     1043,
					HasDefault: false,
				},
				{
					Mode:       "table",
					Name:       "id",
					TypeId:     23,
					HasDefault: false,
				},
				{
					Mode:       "table",
					Name:       "created_at",
					TypeId:     23,
					HasDefault: false,
				},
				{
					Mode:       "table",
					Name:       "sc_name",
					TypeId:     23,
					HasDefault: false,
				},
				{
					Mode:       "table",
					Name:       "c_name",
					TypeId:     23,
					HasDefault: false,
				},
			},
			ArgumentTypes:          "in_scouter_name character varying, in_candidate_name character varying",
			IdentityArgumentTypes:  "in_scouter_name character varying, in_candidate_name character varying",
			ReturnTypeID:           2249,
			ReturnType:             "TABLE(id integer, created_at timestamp without time zone, sc_name character varying, c_name character varying)",
			ReturnTypeRelationID:   0,
			IsSetReturningFunction: true,
			Behavior:               string(raiden.RpcBehaviorVolatile),
			SecurityDefiner:        false,
			ConfigParams:           nil,
		},
	}

	dir, err := os.MkdirTemp("", "rpc")
	assert.NoError(t, err)

	rpcPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(rpcPath)
	assert.NoError(t, err1)

	err2 := generator.GenerateRpc(dir, "test", fns, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.FileExists(t, dir+"/internal/rpc/get_submissions.go")
}
