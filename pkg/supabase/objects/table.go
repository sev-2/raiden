package objects

// ----- table structure definitions -----

type ReplicaIdentity string

const (
	ReplicaIdentityDefault ReplicaIdentity = "DEFAULT"
	ReplicaIdentityIndex   ReplicaIdentity = "INDEX"
)

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
	// TODO : implement check and comment in models
	Check   any `json:"check"`
	Comment any `json:"comment"`
}

type PrimaryKey struct {
	Name      string `json:"name"`
	Schema    string `json:"schema"`
	TableID   int    `json:"table_id"`
	TableName string `json:"table_name"`
}

type Table struct {
	Bytes            int                  `json:"bytes"`
	Columns          []Column             `json:"columns"`
	Comment          any                  `json:"comment"`
	DeadRowsEstimate int                  `json:"dead_rows_estimate"`
	ID               int                  `json:"id"`
	LiveRowsEstimate int                  `json:"live_rows_estimate"`
	Name             string               `json:"name"`
	PrimaryKeys      []PrimaryKey         `json:"primary_keys"`
	Relationships    []TablesRelationship `json:"relationships"`
	ReplicaIdentity  ReplicaIdentity      `json:"replica_identity"`
	RLSEnabled       bool                 `json:"rls_enabled"`
	RLSForced        bool                 `json:"rls_forced"`
	Schema           string               `json:"schema"`
	Size             string               `json:"size"`
}

// ---- update table struct definitions ----
type UpdateTableType string
type UpdateColumnType string

type UpdateColumnItem struct {
	Name        string
	UpdateItems []UpdateColumnType
}

const (
	UpdateTableSchema          UpdateTableType = "schema"
	UpdateTableName            UpdateTableType = "name"
	UpdateTableRlsEnable       UpdateTableType = "rls_enable"
	UpdateTableRlsForced       UpdateTableType = "rls_forced"
	UpdateTablePrimaryKey      UpdateTableType = "primary_key"
	UpdateTableReplicaIdentity UpdateTableType = "replica_identity"
)

const (
	UpdateColumnNew          UpdateColumnType = "new"
	UpdateColumnDelete       UpdateColumnType = "delete"
	UpdateColumnName         UpdateColumnType = "name"
	UpdateColumnDefaultValue UpdateColumnType = "default_value"
	UpdateColumnDataType     UpdateColumnType = "data_type"
	UpdateColumnUnique       UpdateColumnType = "unique"
	UpdateColumnNullable     UpdateColumnType = "nullable"
	UpdateColumnIdentity     UpdateColumnType = "identity"
)

type UpdateTableParam struct {
	OldData           Table
	ChangeColumnItems []UpdateColumnItem
	ChangeItems       []UpdateTableType
}
