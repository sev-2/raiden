package raiden_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestJsonPresenterWriteData(t *testing.T) {
	response := &fasthttp.Response{}

	presenter := raiden.NewJsonPresenter(response)
	data := map[string]interface{}{"message": "test"}
	presenter.SetData(data)

	err := presenter.WriteData()

	assert.NoError(t, err)
	assert.Equal(t, fasthttp.StatusOK, response.StatusCode())

	expectedJSON, _ := json.Marshal(data)
	assert.Equal(t, expectedJSON, response.Body())
}

func TestJsonPresenterWriteError(t *testing.T) {
	response := &fasthttp.Response{}

	presenter := raiden.NewJsonPresenter(response)
	err := errors.New("test error")

	presenter.SetError(err)
	presenter.WriteError(err)

	assert.Equal(t, fasthttp.StatusInternalServerError, response.StatusCode())
	assert.Equal(t, err.Error(), string(response.Body()))
}

func TestJsonPresenterWriteCustomError(t *testing.T) {
	response := &fasthttp.Response{}

	presenter := raiden.NewJsonPresenter(response)
	err := &raiden.ErrorResponse{
		StatusCode: fasthttp.StatusBadRequest,
		Code:       fasthttp.StatusMessage(fasthttp.StatusBadRequest),
		Details:    "invalid request parameter",
		Hint:       "invalid request",
		Message:    "name parameter must be string type",
	}

	presenter.SetError(err)
	presenter.WriteError(err)

	errByte, errMarshall := json.Marshal(err)
	assert.NoError(t, errMarshall)

	assert.Equal(t, fasthttp.StatusBadRequest, response.StatusCode())
	assert.Equal(t, string(errByte), string(response.Body()))
}

func TestJsonPresenterWriteWithErrScenario(t *testing.T) {
	response := &fasthttp.Response{}

	presenter := raiden.NewJsonPresenter(response)

	data := map[string]interface{}{"message": "test"}
	err := errors.New("test error")
	presenter.SetData(data)
	presenter.SetError(err)

	// Call Write method
	presenter.Write()

	assert.Equal(t, "application/json", string(response.Header.ContentType()))
	assert.Equal(t, fasthttp.StatusInternalServerError, response.StatusCode())
	assert.Equal(t, err.Error(), string(response.Body()))
}

func TestJsonPresenterWriteWithSuccessScenario(t *testing.T) {
	response := &fasthttp.Response{}

	presenter := raiden.NewJsonPresenter(response)

	data := map[string]interface{}{"message": "test"}
	presenter.SetData(data)

	// Call Write method
	presenter.Write()

	responseByte, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Equal(t, "application/json", string(response.Header.ContentType()))
	assert.Equal(t, fasthttp.StatusOK, response.StatusCode())
	assert.Equal(t, string(responseByte), string(response.Body()))
}
