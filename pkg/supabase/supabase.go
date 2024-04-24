package supabase

import (
	"errors"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/drivers/cloud"
	"github.com/sev-2/raiden/pkg/supabase/drivers/cloud/admin"
	"github.com/sev-2/raiden/pkg/supabase/drivers/local/meta"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

var SupabaseLogger = logger.HcLog().Named("supabase")

var (
	DefaultApiUrl         = "https://api.supabase.com"
	DefaultIncludedSchema = []string{"public", "auth"}
)

func FindProject(cfg *raiden.Config) (objects.Project, error) {
	return cloud.FindProject(cfg)
}

func GetTables(cfg *raiden.Config, includedSchemas []string) ([]objects.Table, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Get all table from supabase cloud", "project-id", cfg.ProjectId)
		return cloud.GetTables(cfg, includedSchemas, true)
	}
	SupabaseLogger.Debug("Get all table from supabase pg-meta")
	return meta.GetTables(cfg, includedSchemas, true)
}

func CreateTable(cfg *raiden.Config, table objects.Table) (rs objects.Table, err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Create new table to supabase cloud", "table", table.Name, "project-id", cfg.ProjectId)
		return cloud.CreateTable(cfg, table)
	}
	SupabaseLogger.Debug("Create new table in supabase pg-meta")
	return meta.CreateTable(cfg, table)
}

func UpdateTable(cfg *raiden.Config, newTable objects.Table, updateItems objects.UpdateTableParam) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Update table in supabase cloud", "name", updateItems.OldData.Name, "project-id", cfg.ProjectId)
		return cloud.UpdateTable(cfg, newTable, updateItems)
	}
	SupabaseLogger.Debug("Update in supabase pg-meta", "name", updateItems.OldData.Name)
	return meta.UpdateTable(cfg, newTable, updateItems)
}

func DeleteTable(cfg *raiden.Config, table objects.Table, cascade bool) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Delete table supabase cloud", "name", table.Name, "project-id", cfg.ProjectId)
		return cloud.DeleteTable(cfg, table, cascade)
	}
	SupabaseLogger.Debug("Delete table in supabase pg-meta", "name", table.Name)
	return meta.DeleteTable(cfg, table, cascade)
}

func GetRoles(cfg *raiden.Config) ([]objects.Role, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Get all roles from supabase cloud", "project-id", cfg.ProjectId)
		return cloud.GetRoles(cfg)
	}
	SupabaseLogger.Debug("Get all roles from supabase pg-meta")
	return meta.GetRoles(cfg)
}

func CreateRole(cfg *raiden.Config, role objects.Role) (objects.Role, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Create role from supabase cloud", "project-id", cfg.ProjectId)
		return cloud.CreateRole(cfg, role)
	}
	SupabaseLogger.Debug("Create role from supabase pg-meta")
	return meta.CreateRole(cfg, role)
}

func UpdateRole(cfg *raiden.Config, new objects.Role, updateItems objects.UpdateRoleParam) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Update role in supabase cloud", "name", updateItems.OldData.Name, "project-id", cfg.ProjectId)
		return cloud.UpdateRole(cfg, new, updateItems)
	}
	SupabaseLogger.Debug("Update role in supabase pg-meta", "name", updateItems.OldData.Name)
	return meta.UpdateRole(cfg, new, updateItems)
}

func DeleteRole(cfg *raiden.Config, old objects.Role) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Delete role in supabase cloud", "name", old.Name, "project-id", cfg.ProjectId)
		return cloud.DeleteRole(cfg, old)
	}
	SupabaseLogger.Debug("Delete role in supabase pg-meta", "name", old.Name)
	return meta.DeleteRole(cfg, old)
}

func GetPolicies(cfg *raiden.Config) (objects.Policies, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Get all policy from supabase cloud", "project-id", cfg.ProjectId)
		return cloud.GetPolicies(cfg)
	}
	SupabaseLogger.Debug("Get all policy from supabase pg-meta")
	return meta.GetPolicies(cfg)
}

func CreatePolicy(cfg *raiden.Config, policy objects.Policy) (objects.Policy, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Create policy from supabase cloud ", "name", policy.Name, "project-id", cfg.ProjectId)
		return cloud.CreatePolicy(cfg, policy)
	}
	SupabaseLogger.Debug("Create policy from supabase pg-meta")
	return meta.CreatePolicy(cfg, policy)
}

func UpdatePolicy(cfg *raiden.Config, new objects.Policy, updateItems objects.UpdatePolicyParam) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Update policy in supabase cloud", "name", updateItems.Name, "project-id", cfg.ProjectId)
		return cloud.UpdatePolicy(cfg, new, updateItems)
	}
	SupabaseLogger.Debug("Update policy in supabase pg-meta", "name", updateItems.Name)
	return meta.UpdatePolicy(cfg, new, updateItems)
}

func DeletePolicy(cfg *raiden.Config, old objects.Policy) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Delete policy in supabase cloud", "name", old.Name, "project-id", cfg.ProjectId)
		return cloud.DeletePolicy(cfg, old)
	}
	SupabaseLogger.Debug("Delete policy in supabase pg-meta", "name", old.Name)
	return meta.DeletePolicy(cfg, old)
}

func GetFunctions(cfg *raiden.Config) ([]objects.Function, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Get all function from supabase cloud", "project-id", cfg.ProjectId)
		return cloud.GetFunctions(cfg)
	}
	SupabaseLogger.Debug("Get all function from supabase pg-meta")
	return meta.GetFunctions(cfg)
}

func CreateFunction(cfg *raiden.Config, fn objects.Function) (objects.Function, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Create function from supabase cloud", "project-id", cfg.ProjectId)
		return cloud.CreateFunction(cfg, fn)
	}
	SupabaseLogger.Debug("Create function from supabase pg-meta")
	return meta.CreateFunction(cfg, fn)
}

func UpdateFunction(cfg *raiden.Config, fn objects.Function) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Update function in supabase cloud", "name", fn.Name, "project-id", cfg.ProjectId)
		return cloud.UpdateFunction(cfg, fn)
	}
	SupabaseLogger.Debug("Update function in supabase pg-meta", "name", fn.Name)
	return meta.UpdateFunction(cfg, fn)
}

func DeleteFunction(cfg *raiden.Config, fn objects.Function) (err error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Delete function in supabase cloud", "name", fn.Name, "project-id", cfg.ProjectId)
		return cloud.DeleteFunction(cfg, fn)
	}
	SupabaseLogger.Debug("Delete function in supabase pg-meta", "name", fn.Name)
	return meta.DeleteFunction(cfg, fn)
}

func AdminUpdateUserData(cfg *raiden.Config, userId string, data objects.User) (objects.User, error) {
	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		SupabaseLogger.Debug("Update user data in supabase cloud", "user-id", userId, "project-id", cfg.ProjectId)
		return admin.UpdateUser(cfg, userId, data)
	}
	SupabaseLogger.Debug("Update user data in self hosted", "user-id", userId)
	return objects.User{}, errors.New("update user data in self hosted in not implemented, stay update :)")
}
