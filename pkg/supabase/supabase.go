package supabase

import (
	"time"
)

var (
	DefaultApiUrl = "https://api.supabase.com"
	apiUrl        string
)

// Table
type Table struct {
	Bytes            int                  `json:"bytes"`
	Columns          []Column             `json:"columns"`
	Comment          string               `json:"comment"`
	DeadRowsEstimate int                  `json:"dead_rows_estimate"`
	ID               int                  `json:"id"`
	LiveRowsEstimate int                  `json:"live_rows_estimate"`
	Name             string               `json:"name"`
	PrimaryKeys      []PrimaryKey         `json:"primary_keys"`
	Relationships    []TablesRelationship `json:"relationships"`
	ReplicaIdentity  string               `json:"replica_identity"`
	RLSEnabled       bool                 `json:"rls_enabled"`
	RLSForced        bool                 `json:"rls_forced"`
	Schema           string               `json:"schema"`
	Size             string               `json:"size"`
}

type TablesRelationship struct {
	Id                int32  `json:"id"`
	ConstraintName    string `json:"constraint_name"`
	SourceSchema      string `json:"source_schema"`
	SourceTableName   string `json:"source_table_name"`
	SourceColumnName  string `json:"source_column_name"`
	TargetTableSchema string `json:"target_table_schema"`
	TargetTableName   string `json:"target_table_name"`
	TargetColumnName  string `json:"target_column_name"`
}

type Column struct {
	TableID            int      `json:"table_id"`
	Schema             string   `json:"schema"`
	Table              string   `json:"table"`
	ID                 string   `json:"id"`
	OrdinalPosition    int      `json:"ordinal_position"`
	Name               string   `json:"name"`
	DefaultValue       any      `json:"default_value"`
	DataType           string   `json:"data_type"`
	Format             string   `json:"format"`
	IsIdentity         bool     `json:"is_identity"`
	IdentityGeneration any      `json:"identity_generation"`
	IsGenerated        bool     `json:"is_generated"`
	IsNullable         bool     `json:"is_nullable"`
	IsUpdatable        bool     `json:"is_updatable"`
	IsUnique           bool     `json:"is_unique"`
	Enums              []string `json:"enums"`
	Check              string   `json:"check"`
	Comment            string   `json:"comment"`
}

type PrimaryKey struct {
	Name      string `json:"name"`
	Schema    string `json:"schema"`
	TableID   int    `json:"table_id"`
	TableName string `json:"table_name"`
}

func GetTables(projectId *string) ([]Table, error) {
	if projectId != nil {
		return getTables(*projectId, true)
	}

	// todo : get from pg meta

	return []Table{}, nil
}

// Role
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

func GetRoles(projectId *string) ([]Role, error) {
	if projectId != nil {
		return getRoles(*projectId)
	}

	// todo : get from pg meta

	return []Role{}, nil
}

// Policies

type Policy struct {
	ID         int           `json:"id"`
	Schema     string        `json:"schema"`
	Table      string        `json:"table"`
	TableID    int           `json:"table_id"`
	Name       string        `json:"name"`
	Action     string        `json:"action"`
	Roles      []string      `json:"roles"`
	Command    PolicyCommand `json:"command"`
	Definition string        `json:"definition"`
	Check      *string       `json:"check"`
}
type PolicyCommand string

const (
	PolicyCommandSelect PolicyCommand = "SELECT"
	PolicyCommandInsert PolicyCommand = "INSERT"
	PolicyCommandUpdate PolicyCommand = "UPDATE"
	PolicyCommandDelete PolicyCommand = "DELETE"
)

type Policies []Policy

func GetPolicies(projectId *string) ([]Policy, error) {
	if projectId != nil {
		return getPolicies(*projectId)
	}

	// todo : get from pg meta

	return []Policy{}, nil
}

func (p *Policies) FilterByTable(table string) Policies {
	var filteredData Policies
	if p == nil {
		return filteredData
	}

	for _, v := range *p {
		if v.Table == table {
			filteredData = append(filteredData, v)
		}
	}

	return filteredData
}
