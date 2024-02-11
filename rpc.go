package raiden

import (
	"fmt"
	"regexp"
	"strings"
)

// ----- Define type, variable and constant -----

type (
	RpcSecurityType string
	RpcBehaviorType string

	RpcParam struct {
		Name    string
		Type    RpcParamDataType
		Default *string
		Value   any
	}

	Rpc interface {
		BindModels()
		GetName() string
		GetSchema() string
		GetSecurity() RpcSecurityType
		GetBehavior() RpcBehaviorType
		GetDefinition() string
	}

	RpcBase struct {
		Name              string
		Schema            string
		Params            []RpcParam
		Definition        string
		SecurityType      RpcSecurityType
		ReturnType        RpcReturnDataType
		Behavior          RpcBehaviorType
		CompleteStatement string
		Models            map[string]any
	}

	RpcParamTag struct {
		Name         string
		Type         string
		DefaultValue string
	}
)

var (
	DefaultParamPrefix = "in_"
	DefaultRpcSchema   = "public"
)

const (
	RpcBehaviorVolatile  RpcBehaviorType = "VOLATILE"
	RpcBehaviorStable    RpcBehaviorType = "STABLE"
	RpcBehaviorImmutable RpcBehaviorType = "IMMUTABLE"

	RpcSecurityTypeDefiner RpcSecurityType = "DEFINER"
	RpcSecurityTypeInvoker RpcSecurityType = "INVOKER"

	RpcTemplate = `CREATE OR REPLACE FUNCTION :function_name(:params) RETURNS :return_type LANGUAGE plpgsql :security AS $function$ :definition $function$`
)

func MarshalRpcParamTag(paramTag *RpcParamTag) (string, error) {
	if paramTag == nil {
		return "", nil
	}

	var tagArr []string

	if paramTag.Name != "" {
		tagArr = append(tagArr, fmt.Sprintf("name:%s", paramTag.Name))
	}

	if paramTag.Type != "" {
		tagArr = append(tagArr, fmt.Sprintf("type:%s", strings.ToLower(paramTag.Type)))
	}

	if paramTag.DefaultValue != "" {
		tagArr = append(tagArr, fmt.Sprintf("default:%s", paramTag.DefaultValue))
	}

	return strings.Join(tagArr, ";"), nil
}

func UnmarshalRpcParamTag(tag string) (RpcParamTag, error) {
	paramTag := RpcParamTag{}

	// Regular expression to match key-value pairs
	re := regexp.MustCompile(`(\w+):([^;]+);?`)

	// Find all matches in the tag string
	matches := re.FindAllStringSubmatch(tag, -1)

	for _, match := range matches {
		key := match[1]
		value := match[2]

		switch key {
		case "name":
			paramTag.Name = value
		case "type":
			pType, err := GetValidRpcParamType(value, true)
			if err != nil {
				return paramTag, nil
			}
			paramTag.Type = string(pType)
		case "default":
			paramTag.DefaultValue = value
		}

	}
	return paramTag, nil
}

// ----- Rpc base functionality -----
func NewRpc(name string) *RpcBase {
	return &RpcBase{
		Name:         name,
		Schema:       DefaultRpcSchema,
		SecurityType: RpcSecurityTypeInvoker,
		Behavior:     RpcBehaviorVolatile,
	}
}

func (r *RpcBase) BindModel(model any, alias string) *RpcBase {
	r.Models[alias] = model
	return r
}

func (r *RpcBase) GetName() string {
	return ""
}

func (r *RpcBase) GetSchema() string {
	return DefaultRpcSchema
}

func (r *RpcBase) GetSecurity() RpcSecurityType {
	return RpcSecurityTypeInvoker
}

func (r *RpcBase) GetBehavior() RpcBehaviorType {
	return RpcBehaviorVolatile
}

func (r *RpcBase) GetDefinition() string {
	Panicf("Rpc definition must be set")
	return ""
}

func (r *RpcBase) Execute(ctx Context, dest any) {}
