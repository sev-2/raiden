package roles

type SupabaseStorageAdmin struct {
	Metadata string `config:"search_path:storage" connectionLimit:"60" inheritRole:"false" isReplicationRole:"false" isSuperuser:"false"`
	Permission string `canBypassRls:"false" canCreateDb:"false" canCreateRole:"true" canLogin:"true"`
}
