package cloud_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/drivers/cloud"
	"github.com/stretchr/testify/assert"
)

func loadCloudConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    "cloud",
		ProjectId:           "test-project-id",
		SupabaseApiBasePath: "http://localhost:8080",
		SupabaseApiUrl:      "http://localhost:8080",
	}
}

func TestGetTableByName(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := cloud.GetTableByName(cfg, "test-table", "test-schema", false)
	assert.Error(t, err)
}
