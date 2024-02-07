package roles

type Authenticated struct {
	Metadata string `config:"statement_timeout:8s" connectionLimit:"60" inheritRole:"true" isReplicationRole:"false" isSuperuser:"false"`
	Permission string `canBypassRls:"false" canCreateDb:"false" canCreateRole:"false" canLogin:"false"`
}
