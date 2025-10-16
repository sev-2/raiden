package db

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestQuery_SetCredential(t *testing.T) {
	// Create a mock context
	ctx := &raiden.Ctx{
		RequestCtx: &fasthttp.RequestCtx{},
	}

	// Create a new query
	q := NewQuery(ctx)

	// Create a mock credential
	cred := Credential{
		Token:  "test-token",
		ApiKey: "test-api-key",
	}

	// Call SetCredential
	result := q.SetCredential(cred)

	// Assert that the query object is returned (chainable)
	assert.Equal(t, q, result)

	// Assert that the credential was set
	assert.Equal(t, cred, q.credential)
}

func TestQuery_Execute(t *testing.T) {
	// Create a new model base
	model := &ModelBase{}

	// Call Execute
	result := model.Execute()

	// Assert that the model is returned
	assert.Equal(t, model, result)
}

func TestQuery_GetQueryURI(t *testing.T) {
	// Create a mock context
	ctx := &raiden.Ctx{
		RequestCtx: &fasthttp.RequestCtx{},
	}

	// Create a mock model
	type MockModel struct {
		ModelBase
		ID   int64  `json:"id" column:"name:id;type:bigint;primaryKey"`
		Name string `json:"name" column:"name:name;type:text"`

		Metadata string `json:"-" schema:"public" tableName:"mock_models"`
	}

	mockModel := MockModel{}

	// Create a new query with model
	q := NewQuery(ctx).Model(mockModel)

	// Call GetQueryURI
	uri := q.GetQueryURI()

	// Assert that we get a string
	assert.IsType(t, "", uri)
}

func TestNewQueryChainability(t *testing.T) {
	// Create a mock context
	ctx := &raiden.Ctx{
		RequestCtx: &fasthttp.RequestCtx{},
	}

	// Create a mock model
	type MockModel struct {
		ModelBase
		ID   int64  `json:"id" column:"name:id;type:bigint;primaryKey"`
		Name string `json:"name" column:"name:name;type:text"`

		Metadata string `json:"-" schema:"public" tableName:"mock_models"`
	}

	mockModel := MockModel{}

	// Test that we can chain methods together
	cred := Credential{
		Token:  "test-token",
		ApiKey: "test-api-key",
	}

	q := NewQuery(ctx).
		Model(mockModel).
		SetCredential(cred).
		AsSystem()

	// Assert that credential was set
	assert.Equal(t, cred, q.credential)
	// Assert that bypass was set
	assert.True(t, q.ByPass)
}
