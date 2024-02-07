package roles

type SupabaseFunctionsAdmin struct {
	Metadata string `config:"search_path:supabase_functions" connectionLimit:"60" inheritRole:"false" isReplicationRole:"false" isSuperuser:"false"`
	Permission string `canBypassRls:"false" canCreateDb:"false" canCreateRole:"true" canLogin:"true"`
}
