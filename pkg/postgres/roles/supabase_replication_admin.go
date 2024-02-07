package roles

type SupabaseReplicationAdmin struct {
	Metadata string `connectionLimit:"60" inheritRole:"true" isReplicationRole:"true" isSuperuser:"false"`
	Permission string `canBypassRls:"false" canCreateDb:"false" canCreateRole:"false" canLogin:"true"`
}
