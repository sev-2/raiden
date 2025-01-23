package pgmeta_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/client/net"
	"github.com/sev-2/raiden/pkg/connector/pgmeta"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the net.Client interface.
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// Mock implementation of the `Get` function
func TestGetTables_Success(t *testing.T) {
	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	resData := []objects.Table{
		{
			Name:        "test_local_table",
			PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
			Columns: []objects.Column{
				{Name: "id", DataType: "uuid"},
				{Name: "name", DataType: "text"},
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
				},
			},
		},
	}

	mockBody, err := json.Marshal(resData)
	assert.NoError(t, err)
	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBuffer(mockBody)),
	}

	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	// Execute the function under test
	cfg := &raiden.Config{
		PgMetaUrl: "http://mock-pg-meta-url",
	}
	includedSchemas := []string{"public"}
	includeColumns := true

	tables, err := pgmeta.GetTables(cfg, includedSchemas, includeColumns)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, tables, 1)
	assert.Equal(t, "test_local_table", tables[0].Name)
}

func TestGetTableByName_Success(t *testing.T) {
	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	resData := []objects.Table{
		{
			Name:        "test_local_table",
			PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
			Columns: []objects.Column{
				{Name: "id", DataType: "uuid"},
				{Name: "name", DataType: "text"},
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
				},
			},
		},
	}

	mockBody, err := json.Marshal(resData)
	assert.NoError(t, err)
	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBuffer(mockBody)),
	}

	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	// Execute the function under test
	cfg := &raiden.Config{
		PgMetaUrl: "http://mock-pg-meta-url",
	}

	table, err := pgmeta.GetTableByName(cfg, resData[0].Name, "public", true)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "test_local_table", table.Name)
}
