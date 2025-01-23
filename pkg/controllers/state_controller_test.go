package controllers_test

import (
	"os"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/controllers"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/resource"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func loadConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    raiden.DeploymentTargetCloud,
		ProjectId:           "test-project-id",
		ProjectName:         "test-project",
		SupabaseApiBasePath: "/v1",
		SupabaseApiUrl:      "http://supabase.cloud.com",
		SupabasePublicUrl:   "http://supabase.cloud.com",
		Mode:                raiden.BffMode,
	}
}

func TestStateController_Post(t *testing.T) {
	flags := &resource.Flags{
		ProjectPath:     "test_project",
		DryRun:          false,
		UpdateStateOnly: true,
	}
	config := loadConfig()
	config.AllowedTables = "*"
	resource.ImportLogger = logger.HcLog().Named("import")

	mockContext := &mock.MockContext{
		RequestCtx: &fasthttp.RequestCtx{},
		ConfigFn: func() *raiden.Config {
			return config
		},
		SendJsonFn: func(data any) error {
			return nil
		},
	}

	mock := &mock.MockSupabase{Cfg: config}
	mock.Activate()
	defer mock.Deactivate()

	dir, errDir := os.MkdirTemp("", "import")
	assert.NoError(t, errDir)
	flags.ProjectPath = dir

	err0 := mock.MockGetBucketsWithExpectedResponse(200, []objects.Bucket{
		{Name: "some_bucket"},
	})
	assert.NoError(t, err0)

	err1 := mock.MockGetTablesWithExpectedResponse(200, []objects.Table{
		{ID: 1, Name: "some_table", Schema: "public"},
	})
	assert.NoError(t, err1)

	err2 := mock.MockGetFunctionsWithExpectedResponse(200, []objects.Function{
		{ID: 1, Schema: "public", Name: "some_function", Definition: "SELECT * FROM some_table;end $function$", ReturnType: "json"},
	})
	assert.NoError(t, err2)

	err3 := mock.MockGetRolesWithExpectedResponse(200, []objects.Role{
		{
			ID:              1,
			ConnectionLimit: 10,
			Name:            "mock_other_role",
			InheritRole:     true,
			CanLogin:        true,
			CanCreateDB:     true,
			CanCreateRole:   true,
			CanBypassRLS:    true,
		},
	})
	assert.NoError(t, err3)

	c := controllers.StateReadyController{
		Result: controllers.StateReadyResponse{},
	}

	err := c.Post(mockContext)
	assert.NoError(t, err)

	// validate state
	localState, err := state.Load()
	assert.NoError(t, err)
	assert.NotNil(t, localState)
	assert.Len(t, localState.Tables, 1)
	assert.Len(t, localState.Roles, 1)
	assert.Len(t, localState.Rpc, 1)
	assert.Len(t, localState.Storage, 1)
}
