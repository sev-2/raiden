package supabase

import (
	"errors"
	"fmt"

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
		return cloud.GetTables(cfg, includedSchemas, true)
	}
	logger.Debug("Get all table from supabase pg-meta")
	return meta.GetTables(cfg, includedSchemas, true)
}

func CreateTable(cfg *raiden.Config, table objects.Table) (rs objects.Table, err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debugf("Create new table %s in supabase cloud for project id : %s", table.Name, cfg.ProjectId)
		return cloud.CreateTable(cfg, table)
	}
	logger.Debug("Create new table in supabase pg-meta")
	return rs, errors.New("create new table in supabase meta is not implemented not, stay update :)")
}

func UpdateTable(cfg *raiden.Config, oldTable objects.Table, newTable objects.Table, updateItems objects.UpdateTableItem) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debugf("Update table %s in supabase cloud for project id : %s", oldTable.Name, cfg.ProjectId)
		return cloud.UpdateTable(cfg, oldTable, newTable, updateItems)
	}
	logger.Debugf("Update table %s in supabase pg-meta", oldTable.Name)
	return fmt.Errorf("update table %s in supabase meta is not implemented not, stay update :)", oldTable.Name)
}

func DeleteTable(cfg *raiden.Config, table objects.Table, cascade bool) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debugf("Delete table %s in supabase cloud for project id : %s", table.Name, cfg.ProjectId)
		return cloud.DeleteTable(cfg, table, cascade)
	}
	logger.Debugf("Delete table %s in supabase pg-meta", table.Name)
	return fmt.Errorf("delete table %s in supabase meta is not implemented not, stay update :)", table.Name)
}

func GetRoles(cfg *raiden.Config) ([]objects.Role, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debug("Get all roles from supabase cloud with project id : ", cfg.ProjectId)
		return cloud.GetRoles(cfg)
	}
	logger.Debug("Get all roles from supabase pg-meta")
	return meta.GetRoles(cfg)
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
