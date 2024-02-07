package roles

type Postgres struct {
	Metadata   string `config:"search_path:\"\\$user\", public, extensions" connectionLimit:"60" inheritRole:"true" isReplicationRole:"true" isSuperuser:"false"`
	Permission string `canBypassRls:"true" canCreateDb:"true" canCreateRole:"true" canLogin:"true"`
}
