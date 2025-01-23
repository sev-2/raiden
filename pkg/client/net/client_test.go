package net_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/sev-2/raiden/pkg/client/net"
	"github.com/sev-2/raiden/pkg/logger"
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

func setupMockLogger() {
	net.Logger = logger.HcLog().Named("test.logger")
}

func TestSendRequest_Success(t *testing.T) {
	setupMockLogger()

	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	mockResponseBody := `{"message":"success"}`
	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(mockResponseBody)),
	}

	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	body := []byte(`{"key":"value"}`)
	rawBody, err := net.SendRequest(http.MethodPost, "http://example.com", body, 0, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, mockResponseBody, string(rawBody))
	mockClient.AssertExpectations(t)
}

func TestSendRequest_RequestInterceptor(t *testing.T) {
	setupMockLogger()

	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`{"message":"success"}`)),
	}

	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	reqInterceptor := func(req *http.Request) error {
		req.Header.Set("X-Custom-Header", "customValue")
		return nil
	}

	body := []byte(`{"key":"value"}`)
	rawBody, err := net.SendRequest(http.MethodPost, "http://example.com", body, 0, reqInterceptor, nil)

	assert.NoError(t, err)
	assert.NotNil(t, rawBody)
	mockClient.AssertExpectations(t)
}

func TestSendRequest_ErrorResponse(t *testing.T) {
	setupMockLogger()

	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	mockResponse := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(bytes.NewBufferString(`{"message":"internal server error"}`)),
	}

	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	body := []byte(`{"key":"value"}`)
	rawBody, err := net.SendRequest(http.MethodPost, "http://example.com", body, 0, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, rawBody)
	mockClient.AssertExpectations(t)
}

func TestPost_Success(t *testing.T) {
	setupMockLogger()

	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	mockResponseBody := `{"message":"success"}`
	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(mockResponseBody)),
	}

	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	body := []byte(`{"key":"value"}`)
	result, err := net.Post[map[string]string]("http://example.com", body, 0, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, "success", result["message"])
	mockClient.AssertExpectations(t)
}

func TestGet_Success(t *testing.T) {
	setupMockLogger()

	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	mockResponseBody := `{"message":"success"}`
	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(mockResponseBody)),
	}

	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	result, err := net.Get[map[string]string]("http://example.com", 0, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, "success", result["message"])
	mockClient.AssertExpectations(t)
}

func TestPatch_Success(t *testing.T) {
	setupMockLogger()

	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	mockResponseBody := `{"message":"patched successfully"}`
	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(mockResponseBody)),
	}

	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	body := []byte(`{"key":"value"}`)
	result, err := net.Patch[map[string]string]("http://example.com", body, 0, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, "patched successfully", result["message"])
	mockClient.AssertExpectations(t)
}

func TestPut_Success(t *testing.T) {
	setupMockLogger()

	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	mockResponseBody := `{"message":"updated successfully"}`
	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(mockResponseBody)),
	}

	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	body := []byte(`{"key":"value"}`)
	result, err := net.Put[map[string]string]("http://example.com", body, 0, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, "updated successfully", result["message"])
	mockClient.AssertExpectations(t)
}

func TestDelete_Success(t *testing.T) {
	setupMockLogger()

	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	mockResponseBody := `{"message":"deleted successfully"}`
	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(mockResponseBody)),
	}

	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	body := []byte(`{"key":"value"}`)
	result, err := net.Delete[map[string]string]("http://example.com", body, 0, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "deleted successfully", result["message"])
	mockClient.AssertExpectations(t)
}

func TestExtractResponseErr_FromResponse(t *testing.T) {
	setupMockLogger()

	// Mocking a response with a non-2xx status code
	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	// Create a response with a 500 Internal Server Error
	mockResponse := &http.Response{
		StatusCode: 500, // Non-2xx status
		Body:       io.NopCloser(bytes.NewBufferString("Internal Server Error")),
	}

	// Invoke the SendRequest method which will trigger extractResponseErr
	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, errors.New("Internal Server Error"))

	_, err := net.SendRequest(http.MethodPost, "http://example.com", nil, 0, nil, nil)

	// The error should contain a message about the invalid HTTP response code
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Internal Server Error")
	}

	mockClient.AssertExpectations(t)
}

func TestExtractResponseErr_FromTimeoutResponse(t *testing.T) {
	setupMockLogger()

	// Mocking a response with a non-2xx status code
	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	// Create a response with a 500 Internal Server Error
	mockResponse := &http.Response{
		StatusCode: 500, // Non-2xx status
		Body:       io.NopCloser(bytes.NewBufferString("Internal Server Error")),
	}

	// Invoke the SendRequest method which will trigger extractResponseErr
	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, http.ErrHandlerTimeout)

	_, err := net.SendRequest(http.MethodPost, "http://example.com", nil, 0, nil, nil)

	// The error should contain a message about the invalid HTTP response code
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "timeout")
	}

	mockClient.AssertExpectations(t)
}

func TestExtractResponseErr_FromConnClose(t *testing.T) {
	setupMockLogger()

	// Mocking a response with a non-2xx status code
	mockClient := new(MockClient)
	net.GetClient = func() net.Client { return mockClient }

	// Create a response with a 500 Internal Server Error
	mockResponse := &http.Response{
		StatusCode: 500, // Non-2xx status
		Body:       io.NopCloser(bytes.NewBufferString("Internal Server Error")),
	}

	// Invoke the SendRequest method which will trigger extractResponseErr
	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, http.ErrServerClosed)

	_, err := net.SendRequest(http.MethodPost, "http://example.com", nil, 0, nil, nil)

	// The error should contain a message about the invalid HTTP response code
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "conn_close")
	}

	mockClient.AssertExpectations(t)
}
