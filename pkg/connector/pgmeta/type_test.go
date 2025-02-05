package pgmeta_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/connector/pgmeta"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestGetTypes(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
	}

	// Mock the response for the GetTableByName call
	mockedTypeResponse := []objects.Type{
		{
			Name:   "test_type_1",
			Schema: "public",
		},
	}

	httpmock.RegisterResponder("GET", "http://example.com/types",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, mockedTypeResponse)
		},
	)

	// Call the function under test
	result, err := pgmeta.GetTypes(cfg, []string{"public"})

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "test_type_1", result[0].Name)
	assert.Equal(t, "public", result[0].Schema)
}

func TestGetTypes_Err(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
	}

	httpmock.RegisterResponder("GET", "http://example.com/types",
		func(req *http.Request) (*http.Response, error) {
			return nil, http.ErrServerClosed
		},
	)

	// Call the function under test
	_, err := pgmeta.GetTypes(cfg, []string{"public"})

	// Assertions
	assert.Error(t, err)
}

func TestCreateType(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
	}

	// Define the new table to be created
	newType := objects.Type{
		Name:   "test_type_1",
		Schema: "public",
	}

	// Mock the response for the GetTypeByName call
	mockedTypesResponse := []objects.Type{newType}
	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			var payload pgmeta.ExecuteQueryParam
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				return httpmock.NewStringResponse(400, "Bad Request"), nil
			}

			assert.Contains(t, payload.Query, "test_type_1")

			if strings.Contains(payload.Query, "select") {
				return httpmock.NewJsonResponse(200, mockedTypesResponse)
			}

			if strings.Contains(payload.Query, "CREATE TYPE") {
				return httpmock.NewJsonResponse(200, mockedTypesResponse[0])
			}

			return httpmock.NewStringResponse(500, "Internal Server Error"), nil
		},
	)

	// Call the function under test
	result, err := pgmeta.CreateType(cfg, newType)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "test_type_1", result.Name)
	assert.Equal(t, "public", result.Schema)
}

func TestUpdateType(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
	}

	// Mock the response for the GetTableByName call
	data1 := objects.Type{
		Name:   "test_type_1",
		Schema: "public",
		Enums:  []string{"test_1", "test_2"},
	}

	data2 := []objects.Type{data1}

	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, data2)
		},
	)

	// Call the function under test
	err := pgmeta.UpdateType(cfg, data1)

	// Assertions
	assert.NoError(t, err)
}

func TestDeleteType(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
	}

	// Mock the response for the GetTableByName call
	data1 := objects.Type{
		Name:   "test_type_1",
		Schema: "public",
		Enums:  []string{"test_1", "test_2"},
	}

	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, map[string]any{
				"message": "success",
			})
		},
	)

	// Call the function under test
	err := pgmeta.DeleteType(cfg, data1)

	// Assertions
	assert.NoError(t, err)
}
