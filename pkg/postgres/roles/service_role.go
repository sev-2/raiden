package roles

type ServiceRole struct {
	Metadata string `connectionLimit:"60" inheritRole:"true" isReplicationRole:"false" isSuperuser:"false"`
	Permission string `canBypassRls:"true" canCreateDb:"false" canCreateRole:"false" canLogin:"false"`
}
