package objects

type Role struct {
	ActiveConnections int            `json:"active_connections"`
	CanBypassRLS      bool           `json:"can_bypass_rls"`
	CanCreateDB       bool           `json:"can_create_db"`
	CanCreateRole     bool           `json:"can_create_role"`
	CanLogin          bool           `json:"can_login"`
	Config            map[string]any `json:"config"`
	ConnectionLimit   int            `json:"connection_limit"`
	ID                int            `json:"id"`
	InheritRole       bool           `json:"inherit_role"`
	IsReplicationRole bool           `json:"is_replication_role"`
	IsSuperuser       bool           `json:"is_superuser"`
	Name              string         `json:"name"`
	Password          string         `json:"password"`
	ValidUntil        *SupabaseTime  `json:"valid_until"`
}

type UpdateRoleType string

const (
	UpdateConnectionLimit   UpdateRoleType = "connection_limit"
	UpdateRoleName          UpdateRoleType = "name"
	UpdateRoleIsReplication UpdateRoleType = "is_replication"
	UpdateRoleIsSuperUser   UpdateRoleType = "is_superuser"
	UpdateRoleInheritRole   UpdateRoleType = "inherit_role"
	UpdateRoleCanCreateDb   UpdateRoleType = "can_create_db"
	UpdateRoleCanCreateRole UpdateRoleType = "can_create_role"
	UpdateRoleCanLogin      UpdateRoleType = "can_login"
	UpdateRoleCanBypassRls  UpdateRoleType = "can_bypass_rls"
	UpdateRoleConfig        UpdateRoleType = "config"
	UpdateRoleValidUntil    UpdateRoleType = "valid_until"
)

type UpdateRoleParam struct {
	OldData     Role
	ChangeItems []UpdateRoleType
}
