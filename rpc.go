package raiden

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/sev-2/raiden/pkg/utils"
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
		GetDefinition() string

		BuildQuery() (q string, hash string)
		Execute(ctx Context, dest any)
	}

	RpcBase struct {
		Schema            string
		Name              string
		Params            []RpcParam
		Definition        string
		SecurityType      RpcSecurityType
		ReturnType        RpcReturnDataType
		Hash              string
		CompleteStatement string
		Models            map[string]any
	}

	RpcParamTag struct {
		Name         string
		Type         string
		DefaultValue string
	}

	RpcMetadataTag struct {
		Name     string
		Schema   string
		Security RpcSecurityType
		Behavior RpcBehaviorType
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
func (r *RpcBase) GetReturnType() RpcReturnDataType {
	return ""
}

func (r *RpcBase) BuildQuery() (q string, hash string) {
	return
}

func (r *RpcBase) Execute(ctx Context, dest any) {}

// ---- Rpc Param Tag -----
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

// ----- Rpc Metadata Tag -----
func MarshalRpcMetadataTag(metadataTag *RpcMetadataTag) (string, error) {
	if metadataTag == nil {
		return "", nil
	}

	var tagArr []string

	if metadataTag.Name != "" {
		tagArr = append(tagArr, fmt.Sprintf("name:%q", metadataTag.Name))
	}

	if metadataTag.Schema != "" {
		tagArr = append(tagArr, fmt.Sprintf("schema:%q", metadataTag.Schema))
	}

	if metadataTag.Security != "" {
		tagArr = append(tagArr, fmt.Sprintf("security:%q", strings.ToLower(string(metadataTag.Security))))
	}

	if metadataTag.Behavior != "" {
		tagArr = append(tagArr, fmt.Sprintf("behavior:%q", strings.ToLower(string(metadataTag.Behavior))))
	}

	return strings.Join(tagArr, " "), nil
}

func UnmarshalRpcMetadataTag(rawTag string) (RpcMetadataTag, error) {
	var metadata RpcMetadataTag
	tagMap := utils.ParseTag(rawTag)

	fnName, isExist := tagMap["name"]
	if !isExist {
		return metadata, errors.New("rpc metadata : name is must be set in metadata tag")
	}
	metadata.Name = fnName

	if fnSchema, isExist := tagMap["schema"]; !isExist {
		metadata.Schema = DefaultRpcSchema
	} else {
		metadata.Schema = fnSchema
	}

	if fnSecurity, isExist := tagMap["security"]; !isExist {
		metadata.Security = RpcSecurityTypeInvoker
	} else {
		switch strings.ToUpper(fnSecurity) {
		case string(RpcSecurityTypeInvoker):
			metadata.Security = RpcSecurityTypeInvoker
		case string(RpcSecurityTypeDefiner):
			metadata.Security = RpcSecurityTypeDefiner
		default:
			return metadata, errors.New("rpc metadata : invalid security tag " + fnSecurity)
		}

	}

	if fnBehavior, isExist := tagMap["volatile"]; !isExist {
		metadata.Behavior = RpcBehaviorVolatile
	} else {
		switch strings.ToUpper(fnBehavior) {
		case string(RpcBehaviorVolatile):
			metadata.Behavior = RpcBehaviorVolatile
		case string(RpcBehaviorStable):
			metadata.Behavior = RpcBehaviorStable
		case string(RpcBehaviorImmutable):
			metadata.Behavior = RpcBehaviorImmutable
		default:
			return metadata, errors.New("rpc metadata : invalid behavior tag  " + fnBehavior)
		}
	}

	return metadata, nil
}
