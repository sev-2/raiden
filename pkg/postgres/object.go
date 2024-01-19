package postgres

import (
	"time"
)

type Role struct {
	ActiveConnections int        `json:"active_connections"`
	CanBypassRLS      bool       `json:"can_bypass_rls"`
	CanCreateDB       bool       `json:"can_create_db"`
	CanCreateRole     bool       `json:"can_create_role"`
	CanLogin          bool       `json:"can_login"`
	Config            []string   `json:"config"`
	ConnectionLimit   int        `json:"connection_limit"`
	ID                int        `json:"id"`
	InheritRole       bool       `json:"inherit_role"`
	IsReplicationRole bool       `json:"is_replication_role"`
	IsSuperuser       bool       `json:"is_superuser"`
	Name              string     `json:"name"`
	Password          string     `json:"password"`
	ValidUntil        *time.Time `json:"valid_until"`
}
