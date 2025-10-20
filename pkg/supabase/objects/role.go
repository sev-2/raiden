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

	// Preload data when import or apply
	InheritRoles []*Role `json:"inherit_roles"`
}

type Roles []Role

func (rr Roles) ToMap() map[int]Role {
	mapRoles := make(map[int]Role)

	for _, r := range rr {
		mapRoles[r.ID] = r
	}

	return mapRoles
}

type RoleMembership struct {
	ParentID    int    `json:"parent_id"`
	ParentRole  string `json:"parent_role"`
	InheritID   int    `json:"inherit_id"`
	InheritRole string `json:"inherit_role"`
}

type RoleMemberships []RoleMembership

func (rms RoleMemberships) GroupByInheritId() map[int][]RoleMembership {
	groupedRoleMemberships := make(map[int][]RoleMembership)
	for _, rm := range rms {
		items, exist := groupedRoleMemberships[rm.InheritID]
		if !exist {
			groupedRoleMemberships[rm.InheritID] = []RoleMembership{rm}
		}
		items = append(items, rm)
		groupedRoleMemberships[rm.InheritID] = items
		continue
	}
	return groupedRoleMemberships
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
	OldData            Role
	ChangeItems        []UpdateRoleType
	ChangeInheritItems []UpdateRoleInheritItem
}

// ---- update role inherit ----

type UpdateRoleInheritType string

const (
	UpdateRoleInheritGrant  UpdateRoleInheritType = "grant"
	UpdateRoleInheritRevoke UpdateRoleInheritType = "revoke"
)

type UpdateRoleInheritItem struct {
	Role Role
	Type UpdateRoleInheritType
}
