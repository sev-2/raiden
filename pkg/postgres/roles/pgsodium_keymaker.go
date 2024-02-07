package roles

type PgsodiumKeymaker struct {
	Metadata string `connectionLimit:"60" inheritRole:"true" isReplicationRole:"false" isSuperuser:"false"`
	Permission string `canBypassRls:"false" canCreateDb:"false" canCreateRole:"false" canLogin:"false"`
}
