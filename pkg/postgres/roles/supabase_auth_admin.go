package roles

type SupabaseAuthAdmin struct {
	Metadata string `config:"idle_in_transaction_session_timeout:60000;search_path:auth" connectionLimit:"60" inheritRole:"false" isReplicationRole:"false" isSuperuser:"false"`
	Permission string `canBypassRls:"false" canCreateDb:"false" canCreateRole:"true" canLogin:"true"`
}
