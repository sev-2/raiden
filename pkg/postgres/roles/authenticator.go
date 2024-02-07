package roles

type Authenticator struct {
	Metadata string `config:"statement_timeout:8s;lock_timeout:8s;session_preload_libraries:supautils, safeupdate" connectionLimit:"60" inheritRole:"false" isReplicationRole:"false" isSuperuser:"false"`
	Permission string `canBypassRls:"false" canCreateDb:"false" canCreateRole:"false" canLogin:"true"`
}
