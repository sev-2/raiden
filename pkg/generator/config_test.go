package generator_test

import (
	"os"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/stretchr/testify/assert"
)

func TestGenerateConfig(t *testing.T) {
	dir, err := os.MkdirTemp("", "config")
	assert.NoError(t, err)

	conf := &raiden.Config{
		DeploymentTarget:    raiden.DeploymentTargetSelfHosted,
		ProjectId:           "test-project-id",
		ProjectName:         "test-project",
		SupabaseApiBasePath: "/v1",
		SupabaseApiUrl:      "http://supabase.local.com",
		SupabasePublicUrl:   "http://supabase.local.com",
		ServerPort:          "8080",
		ServerHost:          "localhost",
	}

	err1 := generator.GenerateConfig(dir, conf, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err1)
	assert.FileExists(t, dir+"/configs/app.yaml")
}
