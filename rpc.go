package raiden

// ----- Define type, variable and constant -----

type (
	RpcSecurityType string
	RpcBehaviorType string
	RpcDataType     string

	RpcParam struct {
		Name    string
		Type    RpcDataType
		Default *string
		Value   any
	}

	Rpc interface {
		BindModels()
		BindParams()
		GetDefinition() string
		GetReturnType() RpcDataType

		BuildQuery() (q string, hash string)
		Execute(ctx Context, dest any)
	}

	RpcBase struct {
		Schema            string
		Name              string
		Params            []RpcParam
		Definition        string
		SecurityType      RpcSecurityType
		ReturnType        RpcDataType
		Hash              string
		CompleteStatement string
		Models            map[string]any
	}

	RpcMetadataTag struct {
		Name     string
		Schema   string
		Security RpcSecurityType
		Behavior RpcBehaviorType
	}
)

var (
	DefaultParamPrefix = "in"
)

const (
	RpcBehaviorVolatile  RpcBehaviorType = "VOLATILE"
	RpcBehaviorStable    RpcBehaviorType = "STABLE"
	RpcBehaviorImmutable RpcBehaviorType = "IMMUTABLE"

	RpcSecurityTypeDefiner RpcSecurityType = "DEFINER"
	RpcSecurityTypeInvoker RpcSecurityType = "INVOKER"

	RpcTemplate = `CREATE OR REPLACE FUNCTION :function_name(:params) RETURNS :return_type LANGUAGE plpgsql :security AS $function$ :definition $function$`
)

// ----- Rpc base functionality -----
func (r *RpcBase) BindModel(model any, alias string) *RpcBase {
	r.Models[alias] = model
	return r
}

func (r *RpcBase) BindModels() {}
func (r *RpcBase) BindParams() {}
func (r *RpcBase) GetDefinition() string {
	return ""
}
func (r *RpcBase) GetReturnType() RpcDataType {
	return ""
}

func (r *RpcBase) BuildQuery() (q string, hash string) {
	return
}

func (r *RpcBase) Execute(ctx Context, dest any) {}
