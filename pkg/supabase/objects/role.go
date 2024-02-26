package objects

import (
	"fmt"
	"strings"
	"time"

	"github.com/sev-2/raiden/pkg/logger"
)

func NewValidUntil(newTime time.Time) *ValidUntil {
	return &ValidUntil{
		Time: newTime,
	}
}

type ValidUntil struct {
	time.Time
}

func (mt *ValidUntil) UnmarshalJSON(b []byte) error {
	layout := "2006-01-02 00:00:00+00"
	dateString := string(b)
	dateString = strings.ReplaceAll(dateString, "\"", "")
	t, err := time.Parse(layout, dateString)
	if err != nil {
		return err
	}
	mt.Time = t
	return nil
}

func (mt ValidUntil) MarshalJSON() ([]byte, error) {
	logger.Debug("marshall error")
	layout := "2006-01-02 00:00:00+00"
	return []byte(fmt.Sprintf(`"%s"`, mt.Time.Format(layout))), nil
}

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
	ValidUntil        *ValidUntil    `json:"valid_until"`
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
