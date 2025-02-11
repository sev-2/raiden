package pgmeta_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/connector/pgmeta"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestCreateTable(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
		JwtToken:  "meta token",
	}

	// Define the new table to be created
	newTable := objects.Table{
		Name:        "test_local_table",
		PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
		Schema:      "public",
		Columns: []objects.Column{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "text"},
		},

		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "test_constraint_old",
				SourceSchema:      "private",
				SourceTableName:   "test_table",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "old_table",
				TargetColumnName:  "id",
			},
			{
				ConstraintName:    "",
				SourceSchema:      "public",
				SourceTableName:   "test_table1",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "old_table1",
				TargetColumnName:  "id",
			},
		},
	}

	// Mock the response for the GetTableByName call
	mockedTableResponse := []objects.Table{newTable}
	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			var payload pgmeta.ExecuteQueryParam
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				return httpmock.NewStringResponse(400, "Bad Request"), nil
			}
			assert.Contains(t, payload.Query, "test_local_table")

			if strings.Contains(payload.Query, "SELECT") {
				return httpmock.NewJsonResponse(200, mockedTableResponse)
			}

			if strings.Contains(payload.Query, "CREATE TABLE") {
				return httpmock.NewJsonResponse(200, mockedTableResponse[0])
			}

			return httpmock.NewStringResponse(500, "Internal Server Error"), nil
		},
	)

	// Call the function under test
	result, err := pgmeta.CreateTable(cfg, newTable)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "test_local_table", result.Name)
	assert.Equal(t, "public", result.Schema)
}

func TestGetTables(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
		JwtToken:  "meta token",
	}

	// Mock the response for the GetTableByName call
	mockedTableResponse := []objects.Table{
		{
			Name:        "test_local_table",
			Schema:      "public",
			PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
			Columns: []objects.Column{
				{Name: "id", DataType: "uuid"},
				{Name: "name", DataType: "text"},
			},
		},
	}
	httpmock.RegisterResponder("GET", "http://example.com/tables",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, mockedTableResponse)

		},
	)

	// Call the function under test
	result, err := pgmeta.GetTables(cfg, []string{"public"}, true)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "test_local_table", result[0].Name)
	assert.Equal(t, "public", result[0].Schema)
}

func TestUpdateTable(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
		JwtToken:  "meta token",
	}

	// Mock the response for the GetTableByName call
	data1 := objects.Table{
		Name:        "test_local_table",
		Schema:      "public",
		PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
		RLSForced:   false,
		Columns: []objects.Column{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "text"},
			{Name: "phone", DataType: "int"},
		},
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "test_local_constraint",
				SourceSchema:      "public",
				SourceTableName:   "test_local_table",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "test_table",
				TargetColumnName:  "id",
				Index: &objects.Index{
					Schema: "public",
					Table:  "test_local_table",
					Name:   "test_local_table_id_index",
				},
			},
		},
	}

	data2 := []objects.Table{data1}

	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, data2)
		},
	)

	// Call the function under test
	err := pgmeta.UpdateTable(cfg, data1, objects.UpdateTableParam{
		OldData:             data1,
		ForceCreateRelation: true,
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableSchema,
			objects.UpdateTableName,
			objects.UpdateTableRlsEnable,
			objects.UpdateTableRlsForced,
			objects.UpdateTablePrimaryKey,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "test_local_constraint",
					SourceSchema:      "public",
					SourceTableName:   "test_local_table",
					SourceColumnName:  "id",
					TargetTableSchema: "public",
					TargetTableName:   "test_table",
					TargetColumnName:  "id",
					Index: &objects.Index{
						Schema: "public",
						Table:  "test_local_table",
						Name:   "test_local_table_id_index",
					},
				},
				Type: objects.UpdateRelationCreate,
			},
		},
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "createAt",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnNew,
				},
			},
			{
				Name: "phone",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnDataType,
					objects.UpdateColumnName,
					objects.UpdateColumnDefaultValue,
					objects.UpdateColumnNullable,
				},
			},
			{
				Name: "name",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnDelete,
				},
			},
		},
	})

	// Assertions
	assert.NoError(t, err)
}

func TestUpdateTable_Err(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
		JwtToken:  "meta token",
	}

	// Mock the response for the GetTableByName call
	data1 := objects.Table{
		Name:        "test_local_table",
		Schema:      "public",
		PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
		RLSForced:   false,
		Columns: []objects.Column{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "text"},
			{Name: "phone", DataType: "int"},
		},
	}

	data2 := []objects.Table{data1}

	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			var payload pgmeta.ExecuteQueryParam
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				return httpmock.NewStringResponse(400, "Bad Request"), nil
			}

			if strings.Contains(payload.Query, "RENAME COLUMN") {
				return nil, errors.New("network error")
			}
			return httpmock.NewJsonResponse(200, data2)
		},
	)

	// Call the function under test
	err := pgmeta.UpdateTable(cfg, data1, objects.UpdateTableParam{
		OldData:             data1,
		ForceCreateRelation: true,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "createAt",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnNew,
				},
			},
			{
				Name: "phone",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnDataType,
					objects.UpdateColumnName,
					objects.UpdateColumnDefaultValue,
					objects.UpdateColumnNullable,
				},
			},
			{
				Name: "name",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnDelete,
				},
			},
		},
	})

	// Assertions
	assert.Error(t, err)
}

func TestUpdateTable_RelationUpdate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
		JwtToken:  "meta token",
	}

	// Mock the response for the GetTableByName call
	data1 := objects.Table{
		Name:        "test_local_table",
		Schema:      "public",
		PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
		RLSForced:   false,
		Columns: []objects.Column{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "text"},
			{Name: "phone", DataType: "int"},
		},
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "test_local_constraint",
				SourceSchema:      "public",
				SourceTableName:   "test_local_table",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "test_table",
				TargetColumnName:  "id",
				Index: &objects.Index{
					Schema: "public",
					Table:  "test_local_table",
					Name:   "test_local_table_id_index",
				},
			},
			{
				ConstraintName:    "test_local_constraint_2",
				SourceSchema:      "public",
				SourceTableName:   "test_local_table_2",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "test_table_2",
				TargetColumnName:  "id",
				Index: &objects.Index{
					Schema: "public",
					Table:  "test_local_table_2",
					Name:   "test_local_table_id_index_2",
				},
			},
			{
				ConstraintName:    "test_local_constraint_new",
				SourceSchema:      "public",
				SourceTableName:   "test_local_table",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "test_table",
				TargetColumnName:  "id",
				Index: &objects.Index{
					Schema: "public",
					Table:  "test_local_table",
					Name:   "test_local_table_id_index",
				},
			},
			{
				ConstraintName:    "existing_constrain_fkey",
				SourceSchema:      "public",
				SourceTableName:   "test_local_table",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "test_table",
				TargetColumnName:  "id",
				Index: &objects.Index{
					Schema: "public",
					Table:  "test_local_table",
					Name:   "test_local_table_id_index",
				},
			},
			{
				ConstraintName:    "",
				SourceSchema:      "public",
				SourceTableName:   "test_exist_table",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "test_table",
				TargetColumnName:  "id",
			},
		},
	}

	data2 := []objects.Table{data1}

	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, data2)
		},
	)

	// Call the function under test
	err := pgmeta.UpdateTable(cfg, data1, objects.UpdateTableParam{
		OldData:             data1,
		ForceCreateRelation: false,
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableSchema,
			objects.UpdateTableName,
			objects.UpdateTableRlsEnable,
			objects.UpdateTableRlsForced,
			objects.UpdateTablePrimaryKey,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationCreate,
			},
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "test_local_constraint",
					SourceSchema:      "public",
					SourceTableName:   "test_local_table_1",
					SourceColumnName:  "id",
					TargetTableSchema: "public",
					TargetTableName:   "test_table_1",
					TargetColumnName:  "id",
					Index: &objects.Index{
						Schema: "public",
						Table:  "test_local_table",
						Name:   "test_local_table_id",
					},
				},
				Type: objects.UpdateRelationUpdate,
			},
			{
				Data: data1.Relationships[1],
				Type: objects.UpdateRelationDelete,
			},
			{
				Data: data1.Relationships[2],
				Type: objects.UpdateRelationCreate,
			},
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "existing_constrain",
					SourceSchema:      "public",
					SourceTableName:   "test_local_table",
					SourceColumnName:  "id",
					TargetTableSchema: "public",
					TargetTableName:   "test_table",
					TargetColumnName:  "id",
					Index: &objects.Index{
						Schema: "public",
						Table:  "test_local_table",
						Name:   "test_local_table_id_index",
					},
				},
				Type: objects.UpdateRelationUpdate,
			},
		},
	})

	// Assertions
	assert.NoError(t, err)
}

func TestDeleteTable(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
		JwtToken:  "meta token",
	}

	// Mock the response for the GetTableByName call
	data1 := objects.Table{
		Name:        "test_local_table",
		Schema:      "public",
		PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
		RLSForced:   false,
		Columns: []objects.Column{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "text"},
		},
	}

	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, map[string]any{
				"message": "success",
			})
		},
	)

	// Call the function under test
	err := pgmeta.DeleteTable(cfg, data1, true)

	// Assertions
	assert.NoError(t, err)
}

func TestRelationshipActions(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
	}

	// Mock the response for the GetTableByName call
	data := []objects.TablesRelationshipAction{
		{
			ConstraintName: "constraint1",
			UpdateAction:   "c",
			DeletionAction: "c",
		},
		{
			ConstraintName: "constraint2",
			UpdateAction:   "c",
			DeletionAction: "c",
		},
	}

	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, data)
		},
	)

	// Call the function under test
	result, err := pgmeta.GetTableRelationshipActions(cfg, "public")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
}

func TestGetIndexes(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
	}

	// Mock the response for the GetTableByName call
	mockedIndexResponse := []objects.Index{
		{Schema: "public", Table: "table1", Name: "index1", Definition: "index1"},
	}
	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, mockedIndexResponse)

		},
	)

	// Call the function under test
	result, err := pgmeta.GetIndexes(cfg, "public")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}
