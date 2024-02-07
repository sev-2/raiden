package roles

type Anon struct {
	Metadata string `config:"statement_timeout:3s" connectionLimit:"60" inheritRole:"true" isReplicationRole:"false" isSuperuser:"false"`
	Permission string `canBypassRls:"false" canCreateDb:"false" canCreateRole:"false" canLogin:"false"`
}
