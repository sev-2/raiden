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

	// Preload data when import or apply
	Action *TablesRelationshipAction `json:"action"`
	Index  *Index                    `json:"index"`
}

// ----- relation action

// ----- relation action map
// a: No action
// r: Restrict
// c: Cascade
// n: Set null
// d: Set default

type RelationAction string
type RelationActionLabel string

const (
	RelationActionNoAction RelationAction = "a"
	RelationActionRestrict RelationAction = "r"
	RelationActionCascade  RelationAction = "c"
	RelationActionSetNull  RelationAction = "n"
	RelationActionDefault  RelationAction = "d"

	RelationActionNoActionLabel RelationActionLabel = "no action"
	RelationActionRestrictLabel RelationActionLabel = "restrict"
	RelationActionCascadeLabel  RelationActionLabel = "cascade"
	RelationActionSetNullLabel  RelationActionLabel = "set null"
	RelationActionDefaultLabel  RelationActionLabel = "set default"
)

var RelationActionMapLabel = map[RelationAction]RelationActionLabel{
	RelationActionNoAction: RelationActionNoActionLabel,
	RelationActionRestrict: RelationActionRestrictLabel,
	RelationActionCascade:  RelationActionCascadeLabel,
	RelationActionSetNull:  RelationActionSetNullLabel,
	RelationActionDefault:  RelationActionDefaultLabel,
}

var RelationActionMapCode = map[RelationActionLabel]RelationAction{
	RelationActionNoActionLabel: RelationActionNoAction,
	RelationActionRestrictLabel: RelationActionRestrict,
	RelationActionCascadeLabel:  RelationActionCascade,
	RelationActionSetNullLabel:  RelationActionSetNull,
	RelationActionDefaultLabel:  RelationActionDefault,
}

type TablesRelationshipAction struct {
	ID             int    `json:"id"`
	ConstraintName string `json:"constraint_name"`
	DeletionAction string `json:"deletion_action"`
	UpdateAction   string `json:"update_action"`
	SourceID       int    `json:"source_id"`
	SourceSchema   string `json:"source_schema"`
	SourceTable    string `json:"source_table"`
	SourceColumns  string `json:"source_columns"`
	TargetID       int    `json:"target_id"`
	TargetSchema   string `json:"target_schema"`
	TargetTable    string `json:"target_table"`
	TargetColumns  string `json:"target_columns"`
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
type UpdateRelationType string

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

const (
	UpdateRelationCreate         UpdateRelationType = "create"
	UpdateRelationUpdate         UpdateRelationType = "update"
	UpdateRelationDelete         UpdateRelationType = "delete"
	UpdateRelationActionOnUpdate UpdateRelationType = "on_update_action"
	UpdateRelationActionOnDelete UpdateRelationType = "on_delete_action"
	UpdateRelationCreateIndex    UpdateRelationType = "index"
)

type UpdateColumnItem struct {
	Name        string
	UpdateItems []UpdateColumnType
}

type UpdateRelationItem struct {
	Data TablesRelationship
	Type UpdateRelationType
}

type UpdateTableParam struct {
	OldData             Table
	ChangeRelationItems []UpdateRelationItem
	ChangeColumnItems   []UpdateColumnItem
	ChangeItems         []UpdateTableType
	ForceCreateRelation bool
}
