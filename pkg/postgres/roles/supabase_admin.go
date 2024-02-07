package roles

type SupabaseAdmin struct {
	Metadata string `config:"search_path:\"$user\", public, auth, extensions" connectionLimit:"60" inheritRole:"true" isReplicationRole:"true" isSuperuser:"true"`
	Permission string `canBypassRls:"true" canCreateDb:"true" canCreateRole:"true" canLogin:"true"`
}
