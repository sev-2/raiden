package pgmeta_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/connector/pgmeta"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

var mockFunctionData = objects.Function{
	ID:                30369,
	Schema:            "public",
	Name:              "get_next_checkout_id",
	Language:          "sql",
	Definition:        "\n    SELECT nextval('checkout_sequence');\n",
	CompleteStatement: "CREATE OR REPLACE FUNCTION public.get_next_checkout_id()\n RETURNS bigint\n LANGUAGE sql\nAS $function$\n    SELECT nextval('checkout_sequence');\n$function$\n",
	ReturnTypeID:      20,
	ReturnType:        "bigint",
	Behavior:          string(raiden.RpcBehaviorVolatile),
	SecurityDefiner:   true,
}

func TestCreateFunction(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
		JwtToken:  "meta token",
	}

	// Define the new table to be created
	newFunction := mockFunctionData

	// Mock the response for the GetTableByName call
	mockedFunctionResponse := []objects.Function{newFunction}
	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {

			bodyByte, _ := io.ReadAll(req.Body)
			var payload pgmeta.ExecuteQueryParam

			if err := json.Unmarshal(bodyByte, &payload); err != nil {
				return httpmock.NewStringResponse(400, "Bad Request"), nil
			}
			assert.Contains(t, payload.Query, "get_next_checkout_id")

			if strings.Contains(payload.Query, "with functions") {
				return httpmock.NewJsonResponse(200, mockedFunctionResponse)
			}

			if strings.Contains(payload.Query, "CREATE OR REPLACE FUNCTION") {
				return httpmock.NewJsonResponse(200, mockedFunctionResponse[0])
			}

			return httpmock.NewStringResponse(500, "Internal Server Error"), nil
		},
	)

	// Call the function under test
	result, err := pgmeta.CreateFunction(cfg, newFunction)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "get_next_checkout_id", result.Name)
	assert.Equal(t, "public", result.Schema)
}

func TestGetFunctions(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
		JwtToken:  "meta token",
	}

	// Mock the response for the GetTableByName call
	mockedFunctionsResponse := []objects.Function{
		mockFunctionData,
	}
	httpmock.RegisterResponder("GET", "http://example.com/functions",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, mockedFunctionsResponse)

		},
	)

	// Call the function under test
	result, err := pgmeta.GetFunctions(cfg)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "get_next_checkout_id", result[0].Name)
	assert.Equal(t, "public", result[0].Schema)
}

func TestUpdateFunction(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
		JwtToken:  "meta token",
	}

	// Mock the response for the GetTableByName call
	originalData := mockFunctionData
	updatedData := mockFunctionData
	updatedData.Name = fmt.Sprintf("%s_updated", originalData.Name)

	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, updatedData)
		},
	)

	// Call the function under test
	err := pgmeta.UpdateFunction(cfg, updatedData)

	// Assertions
	assert.NoError(t, err)
}

func TestDeleteFunction(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cfg := &raiden.Config{
		PgMetaUrl: "http://example.com",
		ProjectId: "test_project",
		JwtToken:  "meta token",
	}

	// Mock the response for the GetTableByName call
	data1 := mockFunctionData
	httpmock.RegisterResponder("POST", "http://example.com/query",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, map[string]any{
				"message": "success",
			})
		},
	)

	// Call the function under test
	err := pgmeta.DeleteFunction(cfg, data1)

	// Assertions
	assert.NoError(t, err)
}
