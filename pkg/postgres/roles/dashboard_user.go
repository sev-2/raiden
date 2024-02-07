package roles

type DashboardUser struct {
	Metadata string `connectionLimit:"60" inheritRole:"true" isReplicationRole:"true" isSuperuser:"false"`
	Permission string `canBypassRls:"false" canCreateDb:"true" canCreateRole:"true" canLogin:"false"`
}
