package postgres_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/postgres/roles"
	"github.com/stretchr/testify/assert"
)

func TestToGoType(t *testing.T) {
	tests := []struct {
		pgType     postgres.DataType
		isNullable bool
		expected   string
	}{
		{postgres.SmallIntType, false, "int16"},
		{postgres.SmallIntType, true, "*int16"},
		{postgres.IntType, false, "int32"},
		{postgres.BigIntType, false, "int64"},
		{postgres.DecimalType, false, "float64"},
		{postgres.VarcharType, false, "string"},
		{postgres.TimestampType, false, "time.Time"},
		{postgres.BooleanType, false, "bool"},
		{postgres.UuidType, false, "uuid.UUID"},
		{postgres.JsonType, false, "interface{}"},
	}

	for _, test := range tests {
		t.Run(string(test.pgType), func(t *testing.T) {
			result := postgres.ToGoType(test.pgType, test.isNullable)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestToPostgresType(t *testing.T) {
	tests := []struct {
		goType   string
		expected postgres.DataType
	}{
		{"int16", postgres.SmallIntType},
		{"int32", postgres.IntType},
		{"int64", postgres.BigIntType},
		{"uint16", postgres.SmallSerialType},
		{"uint32", postgres.SerialType},
		{"uint64", postgres.BigSerialType},
		{"float32", postgres.RealType},
		{"float64", postgres.DoublePrecisionType},
		{"string", postgres.TextType},
		{"time.Time", postgres.TimestampTzType},
		{"time.Duration", postgres.IntervalType},
		{"bool", postgres.BooleanType},
		{"uuid.UUID", postgres.UuidType},
		{"interface{}", postgres.TextType},
		{"any", postgres.TextType},
	}

	for _, test := range tests {
		t.Run(test.goType, func(t *testing.T) {
			result := postgres.ToPostgresType(test.goType)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsValidDataType(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"smallint", true},
		{"integer", true},
		{"bigint", true},
		{"text", true},
		{"unknown", false},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			result := postgres.IsValidDataType(test.value)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetPgDataTypeName(t *testing.T) {
	tests := []struct {
		pgType      postgres.DataType
		returnAlias bool
		expected    postgres.DataType
	}{
		{postgres.SmallIntType, false, postgres.SmallIntType},
		{postgres.SerialType, false, postgres.SerialType},
		{postgres.IntType, false, postgres.IntType},
		{postgres.DoublePrecisionType, true, postgres.DoublePrecisionTypeAlias},
		{postgres.VarcharType, true, postgres.VarcharTypeAlias},
		{postgres.TimestampTzType, true, postgres.TimestampTzTypeAlias},
		{postgres.TimeTzType, true, postgres.TimeTzTypeAlias},
		{postgres.JsonbType, false, postgres.JsonType},
	}

	for _, test := range tests {
		t.Run(string(test.pgType), func(t *testing.T) {
			result := postgres.GetPgDataTypeName(test.pgType, test.returnAlias)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestKnownRoles(t *testing.T) {
	tests := []struct {
		roleName  string
		knownRole raiden.Role
	}{
		{"anon", &roles.Anon{}},
		{"authenticated", &roles.Authenticated{}},
		{"authenticator", &roles.Authenticator{}},
		{"dashboard_user", &roles.DashboardUser{}},
		{"pg_checkpoint", &roles.PgCheckpoint{}},
		{"pg_database_owner", &roles.PgDatabaseOwner{}},
		{"pg_execute_server_program", &roles.PgExecuteServerProgram{}},
		{"pg_monitor", &roles.PgMonitor{}},
		{"pg_read_all_data", &roles.PgReadAllData{}},
		{"pg_read_all_settings", &roles.PgReadAllSettings{}},
		{"pg_read_all_stats", &roles.PgReadAllStats{}},
		{"pg_read_server_files", &roles.PgReadServerFiles{}},
		{"pg_signal_backend", &roles.PgSignalBackend{}},
		{"pg_stat_scan_tables", &roles.PgStatScanTables{}},
		{"pg_write_all_data", &roles.PgWriteAllData{}},
		{"pg_write_server_files", &roles.PgWriteServerFiles{}},
		{"pgbouncer", &roles.Pgbouncer{}},
		{"pgsodium_keyholder", &roles.PgsodiumKeyholder{}},
		{"pgsodium_keyiduser", &roles.PgsodiumKeyiduser{}},
		{"pgsodium_keymaker", &roles.PgsodiumKeymaker{}},
		{"postgres", &roles.Postgres{}},
		{"service_role", &roles.ServiceRole{}},
		{"supabase_admin", &roles.SupabaseAdmin{}},
		{"supabase_auth_admin", &roles.SupabaseAuthAdmin{}},
		{"supabase_functions_admin", &roles.SupabaseFunctionsAdmin{}},
		{"supabase_read_only_user", &roles.SupabaseReadOnlyUser{}},
		{"supabase_realtime_admin", &roles.SupabaseRealtimeAdmin{}},
		{"supabase_replication_admin", &roles.SupabaseReplicationAdmin{}},
		{"supabase_storage_admin", &roles.SupabaseStorageAdmin{}},
	}

	for _, test := range tests {
		t.Run(test.roleName, func(t *testing.T) {
			assert.Equal(t, test.roleName, test.knownRole.Name())
		})
	}
}
