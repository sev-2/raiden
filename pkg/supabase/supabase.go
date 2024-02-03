package supabase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/sev-2/raiden/pkg/logger"
)

var (
	DefaultApiUrl         = "https://api.supabase.com"
	DefaultIncludedSchema = []string{"public", "auth"}
	apiUrl                string
)

// ----- Table Functionality -----
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
	ReplicaIdentity  any                  `json:"replica_identity"`
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
	Check              any      `json:"check"`
	Comment            any      `json:"comment"`
}

type PrimaryKey struct {
	Name      string `json:"name"`
	Schema    string `json:"schema"`
	TableID   int    `json:"table_id"`
	TableName string `json:"table_name"`
}

func GetTables(ctx context.Context, projectId *string) ([]Table, error) {
	if projectId != nil && *projectId != "" {
		logger.Debug("Get all table from supabase cloud with project id : ", *projectId)
		return getTables(ctx, *projectId, true)
	}
	logger.Debug("Get all table from supabase pg-meta")
	return metaGetTables(ctx, DefaultIncludedSchema, true)
}

// ----- Role Functionality -----
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
	ValidUntil        *time.Time     `json:"valid_until"`
}

func GetRoles(ctx context.Context, projectId *string) ([]Role, error) {
	if projectId != nil && *projectId != "" {
		logger.Debug("Get all roles from supabase cloud with project id : ", *projectId)
		return getRoles(ctx, *projectId)
	}
	logger.Debug("Get all roles from supabase pg-meta")
	return metaGetRoles(ctx)
}

// ----- Policies Functionality -----

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

func GetPolicies(ctx context.Context, projectId *string) (Policies, error) {
	if projectId != nil && *projectId != "" {
		logger.Debug("Get all policy from supabase cloud with project id : ", *projectId)
		return getPolicies(ctx, *projectId)
	}
	logger.Debug("Get all policy from supabase pg-meta")
	return metaGetPolicies(ctx)
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

// ----- Function Functionality -----
type FunctionArg struct {
	Mode       string `json:"mode"`
	Name       string `json:"name"`
	TypeId     int    `json:"type_id"`
	HasDefault bool   `json:"has_default"`
}

type Function struct {
	ID                     int           `json:"id"`
	Schema                 string        `json:"schema"`
	Name                   string        `json:"name"`
	Language               string        `json:"language"`
	Definition             string        `json:"definition"`
	CompleteStatement      string        `json:"complete_statement"`
	Args                   []FunctionArg `json:"args"`
	ArgumentTypes          string        `json:"argument_types"`
	IdentityArgumentTypes  string        `json:"identity_argument_types"`
	ReturnTypeID           int           `json:"return_type_id"`
	ReturnType             string        `json:"return_type"`
	ReturnTypeRelationID   int           `json:"return_type_relation_id"`
	IsSetReturningFunction bool          `json:"is_set_returning_function"`
	Behavior               string        `json:"behavior"`
	SecurityDefiner        bool          `json:"security_definer"`
	ConfigParams           any           `json:"config_params"`
}

func GetFunctions(ctx context.Context, projectId *string) ([]Function, error) {
	if projectId != nil && *projectId != "" {
		logger.Debug("Get all function from supabase cloud with project id : ", *projectId)
		return getFunctions(ctx, *projectId)
	}
	logger.Debug("Get all function from supabase pg-meta")
	return metaGetFunctions(ctx)
}

// ----- execute query -----
func ExecQuery(ctx context.Context, projectId *string, query string) (rs any, err error) {
	if projectId != nil && *projectId != "" {
		logger.Debug("exec query to supabase cloud with project id : ", *projectId)
		return executeQuery[any](ctx, *projectId, "exec query", query, nil)
	}
	return metaExecQuery(context.Background(), query)
}

// ----- logger http client -----
var LoggerHttpClient = http.Client{
	Transport: &LoggerHttpTransport{
		Transport: http.DefaultTransport,
	},
}

type LoggerHttpTransport struct {
	Transport http.RoundTripper
}

func (c *LoggerHttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Print the request details
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	logger.Info("Request:")
	logger.Info(string(dump))

	// Use the original transport to perform the actual HTTP round trip
	resp, err := c.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Print the response details
	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	logger.Info("Response:")
	logger.Info(string(dump))

	return resp, nil
}

// helper function
func marshallResponse[T any](action string, data any, response *http.Response) (result T, err error) {
	if response.StatusCode != http.StatusOK {
		defaultError := fmt.Sprintf("failed get all table with response code : %v", response.StatusCode)
		err = errors.New(defaultError)
		return
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		err = fmt.Errorf("invalid marshall data for %s table : %v", action, err)
		return
	}

	if err = json.Unmarshal(jsonStr, &result); err != nil {
		err = fmt.Errorf("invalid response from %s table : %v", action, err)
		return
	}

	return
}
