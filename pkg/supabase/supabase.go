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
	return rs, errors.New("create new table in supabase meta is not implemented now, stay update :)")
}

func UpdateTable(cfg *raiden.Config, newTable objects.Table, updateItems objects.UpdateTableParam) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debugf("Update table %s in supabase cloud for project id : %s", updateItems.OldData.Name, cfg.ProjectId)
		return cloud.UpdateTable(cfg, newTable, updateItems)
	}
	logger.Debugf("Update table %s in supabase pg-meta", updateItems.OldData.Name)
	return fmt.Errorf("update table %s in supabase meta is not implemented now, stay update :)", updateItems.OldData.Name)
}

func DeleteTable(cfg *raiden.Config, table objects.Table, cascade bool) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debugf("Delete table %s in supabase cloud for project id : %s", table.Name, cfg.ProjectId)
		return cloud.DeleteTable(cfg, table, cascade)
	}
	logger.Debugf("Delete table %s in supabase pg-meta", table.Name)
	return fmt.Errorf("delete table %s in supabase meta is not implemented now, stay update :)", table.Name)
}

func GetRoles(cfg *raiden.Config) ([]objects.Role, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debug("Get all roles from supabase cloud with project id : ", cfg.ProjectId)
		return cloud.GetRoles(cfg)
	}
	logger.Debug("Get all roles from supabase pg-meta")
	return meta.GetRoles(cfg)
}

func CreateRole(cfg *raiden.Config, role objects.Role) (objects.Role, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debug("Create role from supabase cloud with project id : ", cfg.ProjectId)
		return cloud.CreateRole(cfg, role)
	}
	logger.Debug("Create role from supabase pg-meta")
	return objects.Role{}, fmt.Errorf("create role %s in supabase meta is not implemented now, stay update :)", role.Name)
}

func UpdateRole(cfg *raiden.Config, new objects.Role, updateItems objects.UpdateRoleParam) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debugf("Update role %s in supabase cloud for project id : %s", updateItems.OldData.Name, cfg.ProjectId)
		return cloud.UpdateRole(cfg, new, updateItems)
	}
	logger.Debugf("Update role %s in supabase pg-meta", updateItems.OldData.Name)
	return fmt.Errorf("update role %s in supabase meta is not implemented now, stay update :)", updateItems.OldData.Name)
}

func DeleteRole(cfg *raiden.Config, old objects.Role) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debugf("Delete role %s in supabase cloud for project id : %s", old.Name, cfg.ProjectId)
		return cloud.DeleteRole(cfg, old)
	}
	logger.Debugf("Delete role %s in supabase pg-meta", old.Name)
	return fmt.Errorf("delete role %s in supabase meta is not implemented now, stay update :)", old.Name)
}

func GetPolicies(cfg *raiden.Config) (objects.Policies, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debug("Get all policy from supabase cloud with project id : ", cfg.ProjectId)
		return cloud.GetPolicies(cfg)
	}
	logger.Debug("Get all policy from supabase pg-meta")
	return meta.GetPolicies(cfg)
}

func CreatePolicy(cfg *raiden.Config, policy objects.Policy) (objects.Policy, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debug("Create policy from supabase cloud with project id : ", cfg.ProjectId)
		return cloud.CreatePolicy(cfg, policy)
	}
	logger.Debug("Create policy from supabase pg-meta")
	return objects.Policy{}, fmt.Errorf("create policy %s in supabase meta is not implemented now, stay update :)", policy.Name)
}

func UpdatePolicy(cfg *raiden.Config, new objects.Policy, updateItems objects.UpdatePolicyParam) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debugf("Update policy %s in supabase cloud for project id : %s", updateItems.Name, cfg.ProjectId)
		return cloud.UpdatePolicy(cfg, new, updateItems)
	}
	logger.Debugf("Update policy %s in supabase pg-meta", updateItems.Name)
	return fmt.Errorf("update policy %s in supabase meta is not implemented now, stay update :)", updateItems.Name)
}

func DeletePolicy(cfg *raiden.Config, old objects.Policy) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debugf("Delete policy %s in supabase cloud for project id : %s", old.Name, cfg.ProjectId)
		return cloud.DeletePolicy(cfg, old)
	}
	logger.Debugf("Delete policy %s in supabase pg-meta", old.Name)
	return fmt.Errorf("delete policy %s in supabase meta is not implemented now, stay update :)", old.Name)
}

func GetFunctions(cfg *raiden.Config) ([]objects.Function, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debug("Get all function from supabase cloud with project id : ", cfg.ProjectId)
		return cloud.GetFunctions(cfg)
	}
	logger.Debug("Get all function from supabase pg-meta")
	return meta.GetFunctions(cfg)
}
