package generator_test

import (
	"os"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/stretchr/testify/assert"
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

func TestGenerateApplyMainFunction(t *testing.T) {
	dir, err := os.MkdirTemp("", "apply")
	assert.NoError(t, err)

	conf := loadConfig()

	err1 := generator.GenerateApplyMainFunction(dir, conf, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err1)
	applyFile := dir + "/cmd/apply/main.go"
	assert.FileExists(t, applyFile)
}
