package supabase

var (
	apiUrl string
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
		return cloudGetTables(*projectId, true)
	}

	return nil, nil
}

// Role
type Roles struct{}
