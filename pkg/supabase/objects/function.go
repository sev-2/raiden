package objects

type FunctionArg struct {
	Mode       string `json:"mode"`
	Name       string `json:"name"`
	TypeId     int    `json:"type_id"`
	HasDefault bool   `json:"has_default"`
}

type Function struct {
	ID                     int           `json:"id"`
	Schema                 string        `json:"schema"`
	Name                   string        `json:"name"`
	Language               string        `json:"language"`
	Definition             string        `json:"definition"`
	CompleteStatement      string        `json:"complete_statement"`
	Args                   []FunctionArg `json:"args"`
	ArgumentTypes          string        `json:"argument_types"`
	IdentityArgumentTypes  string        `json:"identity_argument_types"`
	ReturnTypeID           int           `json:"return_type_id"`
	ReturnType             string        `json:"return_type"`
	ReturnTypeRelationID   int           `json:"return_type_relation_id"`
	IsSetReturningFunction bool          `json:"is_set_returning_function"`
	Behavior               string        `json:"behavior"`
	SecurityDefiner        bool          `json:"security_definer"`
	ConfigParams           any           `json:"config_params"`
}
