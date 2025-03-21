package db_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/db"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestPostgrestRequest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	sampleConfigFile, err := utils.CreateFile(currentDir+"/app.yaml", true)
	assert.NoError(t, err)
	defer func() {
		err := utils.DeleteFile(currentDir + "/app.yaml")
		assert.NoError(t, err)
	}()

	configContent := `MODE: svc
POSTGREST_URL: http://test.com:3000
JWT_TOKEN: Bearer xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
`
	_, err = sampleConfigFile.WriteString(configContent)
	assert.NoError(t, err)
	sampleConfigFile.Close()

	httpmock.RegisterResponder("POST", "http://test.com/member",
		func(req *http.Request) (*http.Response, error) {
			result := map[string]string{
				"message": "success",
			}
			return httpmock.NewJsonResponse(200, result)
		},
	)

	params := map[string]string{
		"name": "bob",
	}

	pBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	result := make(map[string]string)
	r, e := db.PostgrestRequest(nil, db.Credential{
		Token:  "Bearer xxxxxxxxxxxxxxxx",
		ApiKey: "xxxxxxxxxxxx",
	}, "POST", "/member", pBytes, nil, false, &result)
	assert.NoError(t, e)
	assert.NotNil(t, r)
}

func TestPostgrestRequest_ByPass(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	sampleConfigFile, err := utils.CreateFile(currentDir+"/app.yaml", true)
	assert.NoError(t, err)
	defer func() {
		err := utils.DeleteFile(currentDir + "/app.yaml")
		assert.NoError(t, err)
	}()

	configContent := `MODE: svc
POSTGREST_URL: http://test.com:3000
JWT_TOKEN: Bearer xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
`
	_, err = sampleConfigFile.WriteString(configContent)
	assert.NoError(t, err)
	sampleConfigFile.Close()

	httpmock.RegisterResponder("POST", "http://test.com/member",
		func(req *http.Request) (*http.Response, error) {
			result := map[string]string{
				"message": "success",
			}
			return httpmock.NewJsonResponse(200, result)
		},
	)

	params := map[string]string{
		"name": "bob",
	}

	pBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	result := make(map[string]string)
	r, e := db.PostgrestRequest(nil, db.Credential{
		Token:  "Bearer xxxxxxxxxxxxxxxx",
		ApiKey: "xxxxxxxxxxxx",
	}, "POST", "/member", pBytes, nil, true, &result)
	assert.NoError(t, e)
	assert.NotNil(t, r)
}

func TestPostgrestRequest_BffMode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	sampleConfigFile, err := utils.CreateFile(currentDir+"/app.yaml", true)
	assert.NoError(t, err)
	defer func() {
		err := utils.DeleteFile(currentDir + "/app.yaml")
		assert.NoError(t, err)
	}()

	configContent := `MODE: bff
SUPABASE_API_URL: https://api.supabase.com
SUPABASE_API_BASE_PATH: 
`
	_, err = sampleConfigFile.WriteString(configContent)
	assert.NoError(t, err)
	sampleConfigFile.Close()

	httpmock.RegisterResponder("POST", "https://api.supabase.com/rest/v1/member",
		func(req *http.Request) (*http.Response, error) {
			result := map[string]string{
				"message": "success",
			}
			return httpmock.NewJsonResponse(200, result)
		},
	)

	params := map[string]string{
		"name": "bob",
	}

	pBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	result := make(map[string]string)
	r, e := db.PostgrestRequest(nil, db.Credential{
		Token:  "xxxxxxxxxxxxxxxx",
		ApiKey: "xxxxxxxxxxxx",
	}, "POST", "/member", pBytes, nil, false, &result)
	assert.NoError(t, e)
	assert.NotNil(t, r)
}

func TestPostgrestRequest_PathNoSlashWithInvalidJson(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	sampleConfigFile, err := utils.CreateFile(currentDir+"/app.yaml", true)
	assert.NoError(t, err)
	defer func() {
		err := utils.DeleteFile(currentDir + "/app.yaml")
		assert.NoError(t, err)
	}()

	configContent := `MODE: bff
SUPABASE_API_URL: https://api.supabase.com
SUPABASE_API_BASE_PATH: 
`
	_, err = sampleConfigFile.WriteString(configContent)
	assert.NoError(t, err)
	sampleConfigFile.Close()

	httpmock.RegisterResponder("POST", "https://api.supabase.com/rest/v1/member",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "invalid json data"), nil
		},
	)

	params := map[string]string{
		"name": "bob",
	}

	pBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	result := make([]any, 0)
	r, e := db.PostgrestRequest(nil, db.Credential{
		Token:  "xxxxxxxxxxxxxxxx",
		ApiKey: "xxxxxxxxxxxx",
	}, "POST", "member", pBytes, nil, true, &result)
	assert.Error(t, e)
	assert.NotNil(t, r)
}

func TestPostgrestRequest_NotAllowedErr(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	params := map[string]string{
		"name": "bob",
	}

	pBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	result := make(map[string]string)
	r, e := db.PostgrestRequest(nil, db.Credential{
		Token:  "Bearer xxxxxxxxxxxxxxxx",
		ApiKey: "xxxxxxxxxxxx",
	}, "ERR", "/member", pBytes, nil, false, &result)
	assert.Error(t, e)
	assert.EqualError(t, e, "method ERR is not allowed")
	assert.Nil(t, r)
}

func TestPostgrestRequestBind(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	sampleConfigFile, err := utils.CreateFile(currentDir+"/app.yaml", true)
	assert.NoError(t, err)
	defer func() {
		err := utils.DeleteFile(currentDir + "/app.yaml")
		assert.NoError(t, err)
	}()

	configContent := `MODE: svc
POSTGREST_URL: http://test.com:3000
`
	_, err = sampleConfigFile.WriteString(configContent)
	assert.NoError(t, err)
	sampleConfigFile.Close()

	httpmock.RegisterResponder("POST", "http://test.com/member",
		func(req *http.Request) (*http.Response, error) {
			result := map[string]string{
				"message": "success",
			}
			return httpmock.NewJsonResponse(200, result)
		},
	)

	appCtx := &mock.MockContext{
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{
				PgMetaUrl: "http://test.com",
				Mode:      raiden.SvcMode,
			}
		},
		RequestContextFn: func() *fasthttp.RequestCtx {
			return &fasthttp.RequestCtx{}
		},
	}
	params := map[string]string{
		"name": "bob",
	}

	pBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	result := make(map[string]string)
	r, e := db.PostgrestRequestBind(appCtx, "POST", "/member", pBytes, nil, false, &result, nil)
	assert.NoError(t, e)
	assert.NotNil(t, r)
}

func TestPostgrestRequestBind_CallbackError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	sampleConfigFile, err := utils.CreateFile(currentDir+"/app.yaml", true)
	assert.NoError(t, err)
	defer func() {
		err := utils.DeleteFile(currentDir + "/app.yaml")
		assert.NoError(t, err)
	}()

	configContent := `MODE: svc
POSTGREST_URL: http://test.com:3000
`
	_, err = sampleConfigFile.WriteString(configContent)
	assert.NoError(t, err)
	sampleConfigFile.Close()

	httpmock.RegisterResponder("POST", "http://test.com/member",
		func(req *http.Request) (*http.Response, error) {
			result := map[string]string{
				"message": "success",
			}
			return httpmock.NewJsonResponse(200, result)
		},
	)

	appCtx := &mock.MockContext{
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{
				PgMetaUrl: "http://test.com",
				Mode:      raiden.SvcMode,
			}
		},
		RequestContextFn: func() *fasthttp.RequestCtx {
			return &fasthttp.RequestCtx{}
		},
	}
	params := map[string]string{
		"name": "bob",
	}

	pBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	result := make(map[string]string)
	r, e := db.PostgrestRequestBind(appCtx, "POST", "/member", pBytes, nil, false, &result, func(code int, data []byte) error {
		return errors.New("test error")
	})
	assert.Error(t, e)
	assert.NotNil(t, r)
}

func TestPostgrestRequestBind_PathNoSlash(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	sampleConfigFile, err := utils.CreateFile(currentDir+"/app.yaml", true)
	assert.NoError(t, err)
	defer func() {
		err := utils.DeleteFile(currentDir + "/app.yaml")
		assert.NoError(t, err)
	}()

	configContent := `MODE: svc
POSTGREST_URL: http://localhost:3000
PG_META_URL: http://localhost:8080
`
	_, err = sampleConfigFile.WriteString(configContent)
	assert.NoError(t, err)
	sampleConfigFile.Close()

	httpmock.RegisterResponder("POST", "http://test.com/member",
		func(req *http.Request) (*http.Response, error) {
			result := map[string]string{
				"message": "success",
			}
			return httpmock.NewJsonResponse(200, result)
		},
	)

	appCtx := &mock.MockContext{
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{
				PgMetaUrl: "http://test.com",
				Mode:      raiden.SvcMode,
			}
		},
		RequestContextFn: func() *fasthttp.RequestCtx {
			return &fasthttp.RequestCtx{}
		},
	}
	params := map[string]string{
		"name": "bob",
	}

	pBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	result := make(map[string]string)
	r, e := db.PostgrestRequestBind(appCtx, "POST", "member", pBytes, nil, false, &result, nil)
	assert.NoError(t, e)
	assert.NotNil(t, r)
}

func TestPostgrestRequestBind_PathNoSlashBff(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	sampleConfigFile, err := utils.CreateFile(currentDir+"/app.yaml", true)
	assert.NoError(t, err)
	defer func() {
		err := utils.DeleteFile(currentDir + "/app.yaml")
		assert.NoError(t, err)
	}()

	configContent := `MODE: bff
POSTGREST_URL: http://localhost:3000
PG_META_URL: http://localhost:8080
`
	_, err = sampleConfigFile.WriteString(configContent)
	assert.NoError(t, err)
	sampleConfigFile.Close()

	httpmock.RegisterResponder("POST", "http://test.com/member",
		func(req *http.Request) (*http.Response, error) {
			result := map[string]string{
				"message": "success",
			}
			return httpmock.NewJsonResponse(200, result)
		},
	)

	appCtx := &mock.MockContext{
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{
				PgMetaUrl: "http://test.com",
				Mode:      raiden.SvcMode,
			}
		},
		RequestContextFn: func() *fasthttp.RequestCtx {
			return &fasthttp.RequestCtx{}
		},
	}
	params := map[string]string{
		"name": "bob",
	}

	pBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	result := make(map[string]string)
	r, e := db.PostgrestRequestBind(appCtx, "POST", "member", pBytes, nil, false, &result, nil)
	assert.NoError(t, e)
	assert.NotNil(t, r)
}

func TestPostgrestRequestBindCredential_CallbackError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	sampleConfigFile, err := utils.CreateFile(currentDir+"/app.yaml", true)
	assert.NoError(t, err)
	defer func() {
		err := utils.DeleteFile(currentDir + "/app.yaml")
		assert.NoError(t, err)
	}()

	configContent := `MODE: svc
POSTGREST_URL: http://test.com:3000
`
	_, err = sampleConfigFile.WriteString(configContent)
	assert.NoError(t, err)
	sampleConfigFile.Close()

	httpmock.RegisterResponder("POST", "http://test.com/member",
		func(req *http.Request) (*http.Response, error) {
			result := map[string]string{
				"message": "success",
			}
			return httpmock.NewJsonResponse(200, result)
		},
	)

	params := map[string]string{
		"name": "bob",
	}

	pBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	result := make(map[string]string)
	r, e := db.PostgrestRequestBindCredential(db.Credential{}, "POST", "/member", pBytes, nil, false, &result, func(code int, data []byte) error {
		return errors.New("test error")
	})
	assert.Error(t, e)
	assert.NotNil(t, r)
}

func TestPostgrestRequestBindCredential_WithJwtToken(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	sampleConfigFile, err := utils.CreateFile(currentDir+"/app.yaml", true)
	assert.NoError(t, err)
	defer func() {
		err := utils.DeleteFile(currentDir + "/app.yaml")
		assert.NoError(t, err)
	}()

	configContent := `MODE: svc
POSTGREST_URL: http://test.com:3000
JWT_TOKEN: Bearer xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
`
	_, err = sampleConfigFile.WriteString(configContent)
	assert.NoError(t, err)
	sampleConfigFile.Close()

	httpmock.RegisterResponder("POST", "http://test.com/member",
		func(req *http.Request) (*http.Response, error) {
			result := map[string]string{
				"message": "success",
			}
			return httpmock.NewJsonResponse(200, result)
		},
	)

	params := map[string]string{
		"name": "bob",
	}

	pBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	result := make(map[string]string)
	r, e := db.PostgrestRequestBindCredential(db.Credential{}, "POST", "/member", pBytes, nil, true, &result, func(code int, data []byte) error {
		return errors.New("test error")
	})
	assert.Error(t, e)
	assert.NotNil(t, r)
}
