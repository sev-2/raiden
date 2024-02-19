package supabase

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/drivers/cloud"
	"github.com/sev-2/raiden/pkg/supabase/drivers/local/meta"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

var (
	DefaultApiUrl         = "https://api.supabase.com"
	DefaultIncludedSchema = []string{"public", "auth"}
)

func FindProject(cfg *raiden.Config) (objects.Project, error) {
	return cloud.FindProject(cfg)
}

func GetTables(cfg *raiden.Config, includedSchemas []string) ([]objects.Table, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debug("Get all table from supabase cloud with project id : ", cfg.ProjectId)
		return cloud.GetTable(cfg, includedSchemas, true)
	}
	logger.Debug("Get all table from supabase pg-meta")
	return meta.GetTable(cfg, includedSchemas, true)
}

func GetRoles(cfg *raiden.Config) ([]objects.Role, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debug("Get all roles from supabase cloud with project id : ", cfg.ProjectId)
		return cloud.GetRole(cfg)
	}
	logger.Debug("Get all roles from supabase pg-meta")
	return meta.GetRole(cfg)
}

func GetPolicies(cfg *raiden.Config) (objects.Policies, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debug("Get all policy from supabase cloud with project id : ", cfg.ProjectId)
		return cloud.GetPolicies(cfg)
	}
	logger.Debug("Get all policy from supabase pg-meta")
	return meta.GetPolicies(cfg)
}

func GetFunctions(cfg *raiden.Config) ([]objects.Function, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debug("Get all function from supabase cloud with project id : ", cfg.ProjectId)
		return cloud.GetFunctions(cfg)
	}
	logger.Debug("Get all function from supabase pg-meta")
	return meta.GetFunctions(cfg)
}
