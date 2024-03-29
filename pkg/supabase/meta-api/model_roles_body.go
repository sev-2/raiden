/*
 * postgres-meta
 *
 * A REST API to manage your Postgres database
 *
 * API version: 0.0.0-automated
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package meta_api

type RolesBody struct {
	Name              string            `json:"name"`
	Password          string            `json:"password,omitempty"`
	InheritRole       bool              `json:"inherit_role,omitempty"`
	CanLogin          bool              `json:"can_login,omitempty"`
	IsSuperuser       bool              `json:"is_superuser,omitempty"`
	CanCreateDb       bool              `json:"can_create_db,omitempty"`
	CanCreateRole     bool              `json:"can_create_role,omitempty"`
	IsReplicationRole bool              `json:"is_replication_role,omitempty"`
	CanBypassRls      bool              `json:"can_bypass_rls,omitempty"`
	ConnectionLimit   int32             `json:"connection_limit,omitempty"`
	MemberOf          []string          `json:"member_of,omitempty"`
	Members           []string          `json:"members,omitempty"`
	Admins            []string          `json:"admins,omitempty"`
	ValidUntil        string            `json:"valid_until,omitempty"`
	Config            map[string]string `json:"config,omitempty"`
}