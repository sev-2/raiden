package roles

import "github.com/sev-2/raiden"

var NativeRoles = []raiden.Role{
	&Anon{},
	&Authenticated{},
	&Authenticator{},
	&DashboardUser{},
	&PgCheckpoint{},
	&PgDatabaseOwner{},
	&PgExecuteServerProgram{},
	&PgMonitor{},
	&PgReadAllData{},
	&PgReadAllSettings{},
	&PgReadAllStats{},
	&PgReadServerFiles{},
	&PgSignalBackend{},
	&PgStatScanTables{},
	&PgWriteAllData{},
	&PgWriteServerFiles{},
	&Pgbouncer{},
	&PgsodiumKeyholder{},
	&PgsodiumKeyiduser{},
	&PgsodiumKeymaker{},
	&Postgres{},
	&ServiceRole{},
	&SupabaseAdmin{},
	&SupabaseAuthAdmin{},
	&SupabaseFunctionsAdmin{},
	&SupabaseReadOnlyUser{},
	&SupabaseReplicationAdmin{},
	&SupabaseStorageAdmin{},
}