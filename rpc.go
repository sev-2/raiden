package raiden

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/client/net"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/valyala/fasthttp"
)

var RpcLogger = logger.HcLog().Named("raiden.rpc")

// ---- Define rpc data type -----
type RpcParamDataType string
type RpcReturnDataType string

// Define constants for rpc input data type
const (
	RpcParamDataTypeInteger          RpcParamDataType = "INTEGER"
	RpcParamDataTypeNumeric          RpcParamDataType = "NUMERIC"
	RpcParamDataTypeBigInt           RpcParamDataType = "BIGINT"
	RpcParamDataTypeReal             RpcParamDataType = "REAL"
	RpcParamDataTypeDoublePreci      RpcParamDataType = "DOUBLE PRECISION"
	RpcParamDataTypeText             RpcParamDataType = "TEXT"
	RpcParamDataTypeVarchar          RpcParamDataType = "CHARACTER VARYING"
	RpcParamDataTypeVarcharAlias     RpcParamDataType = "VARCHAR"
	RpcParamDataTypeBoolean          RpcParamDataType = "BOOLEAN"
	RpcParamDataTypeBytea            RpcParamDataType = "BYTEA"
	RpcParamDataTypeTimestamp        RpcParamDataType = "TIMESTAMP WITHOUT TIME ZONE"
	RpcParamDataTypeTimestampAlias   RpcParamDataType = "TIMESTAMP"
	RpcParamDataTypeTimestampTZ      RpcParamDataType = "TIMESTAMP WITH TIME ZONE"
	RpcParamDataTypeTimestampTZAlias RpcParamDataType = "TIMESTAMPZ"
	RpcParamDataTypeJSON             RpcParamDataType = "JSON"
	RpcParamDataTypeJSONB            RpcParamDataType = "JSONB"

	// ARRAY TYPES
	RpcParamDataTypeArrayInteger          RpcParamDataType = "INTEGER[]"
	RpcParamDataTypeArrayBigInt           RpcParamDataType = "BIGINT[]"
	RpcParamDataTypeArrayReal             RpcParamDataType = "REAL[]"
	RpcParamDataTypeArrayDoublePreci      RpcParamDataType = "DOUBLE PRECISION[]"
	RpcParamDataTypeArrayNumeric          RpcParamDataType = "NUMERIC[]"
	RpcParamDataTypeArrayText             RpcParamDataType = "TEXT[]"
	RpcParamDataTypeArrayVarchar          RpcParamDataType = "CHARACTER VARYING[]"
	RpcParamDataTypeArrayVarcharAlias     RpcParamDataType = "VARCHAR[]"
	RpcParamDataTypeArrayBoolean          RpcParamDataType = "BOOLEAN[]"
	RpcParamDataTypeArrayBytea            RpcParamDataType = "BYTEA[]"
	RpcParamDataTypeArrayTimestamp        RpcParamDataType = "TIMESTAMP WITHOUT TIME ZONE[]"
	RpcParamDataTypeArrayTimestampAlias   RpcParamDataType = "TIMESTAMP[]"
	RpcParamDataTypeArrayTimestampTZ      RpcParamDataType = "TIMESTAMP WITH TIME ZONE[]"
	RpcParamDataTypeArrayTimestampTZAlias RpcParamDataType = "TIMESTAMPZ[]"
	RpcParamDataTypeArrayJSON             RpcParamDataType = "JSON[]"
	RpcParamDataTypeArrayJSONB            RpcParamDataType = "JSONB[]"
)

// Define constants for rpc return data type
const (
	RpcReturnDataTypeInteger          RpcReturnDataType = "INTEGER"
	RpcReturnDataTypeBigInt           RpcReturnDataType = "BIGINT"
	RpcReturnDataTypeReal             RpcReturnDataType = "REAL"
	RpcReturnDataTypeDoublePreci      RpcReturnDataType = "DOUBLE PRECISION"
	RpcReturnDataTypeNumeric          RpcReturnDataType = "NUMERIC"
	RpcReturnDataTypeText             RpcReturnDataType = "TEXT"
	RpcReturnDataTypeVarchar          RpcReturnDataType = "CHARACTER VARYING"
	RpcReturnDataTypeVarcharAlias     RpcReturnDataType = "VARCHAR"
	RpcReturnDataTypeBoolean          RpcReturnDataType = "BOOLEAN"
	RpcReturnDataTypeBytea            RpcReturnDataType = "BYTEA"
	RpcReturnDataTypeTimestamp        RpcReturnDataType = "TIMESTAMP WITHOUT TIME ZONE"
	RpcReturnDataTypeTimestampAlias   RpcReturnDataType = "TIMESTAMP"
	RpcReturnDataTypeTimestampTZ      RpcReturnDataType = "TIMESTAMP WITH TIME ZONE"
	RpcReturnDataTypeTimestampTZAlias RpcReturnDataType = "TIMESTAMPZ"
	RpcReturnDataTypeJSON             RpcReturnDataType = "JSON"
	RpcReturnDataTypeJSONB            RpcReturnDataType = "JSONB"
	RpcReturnDataTypeRecord           RpcReturnDataType = "RECORD" // like tuple
	RpcReturnDataTypeTable            RpcReturnDataType = "TABLE"
	RpcReturnDataTypeSetOf            RpcReturnDataType = "SETOF"
	RpcReturnDataTypeVoid             RpcReturnDataType = "VOID"

	// ARRAY TYPES
	RpcReturnDataTypeArrayInteger          RpcReturnDataType = "INTEGER[]"
	RpcReturnDataTypeArrayBigInt           RpcReturnDataType = "BIGINT[]"
	RpcReturnDataTypeArrayReal             RpcReturnDataType = "REAL[]"
	RpcReturnDataTypeArrayDoublePreci      RpcReturnDataType = "DOUBLE PRECISION[]"
	RpcReturnDataTypeArrayNumeric          RpcReturnDataType = "NUMERIC[]"
	RpcReturnDataTypeArrayText             RpcReturnDataType = "TEXT[]"
	RpcReturnDataTypeArrayVarchar          RpcReturnDataType = "CHARACTER VARYING[]"
	RpcReturnDataTypeArrayVarcharAlias     RpcReturnDataType = "VARCHAR[]"
	RpcReturnDataTypeArrayBoolean          RpcReturnDataType = "BOOLEAN[]"
	RpcReturnDataTypeArrayBytea            RpcReturnDataType = "BYTEA[]"
	RpcReturnDataTypeArrayTimestamp        RpcReturnDataType = "TIMESTAMP WITHOUT TIME ZONE[]"
	RpcReturnDataTypeArrayTimestampAlias   RpcReturnDataType = "TIMESTAMP[]"
	RpcReturnDataTypeArrayTimestampTZ      RpcReturnDataType = "TIMESTAMP WITH TIME ZONE[]"
	RpcReturnDataTypeArrayTimestampTZAlias RpcReturnDataType = "TIMESTAMPZ[]"
	RpcReturnDataTypeArrayJSON             RpcReturnDataType = "JSON[]"
	RpcReturnDataTypeArrayJSONB            RpcReturnDataType = "JSONB[]"
)

func RpcParamToGoType(dataType RpcParamDataType) string {
	switch dataType {
	case RpcParamDataTypeInteger, RpcParamDataTypeBigInt:
		return "int64"
	case RpcParamDataTypeReal:
		return "float32"
	case RpcParamDataTypeDoublePreci, RpcParamDataTypeNumeric:
		return "float64"
	case RpcParamDataTypeText, RpcParamDataTypeVarchar, RpcParamDataTypeVarcharAlias:
		return "string"
	case RpcParamDataTypeBoolean:
		return "bool"
	case RpcParamDataTypeBytea:
		return "[]byte"
	case RpcParamDataTypeTimestamp, RpcParamDataTypeTimestampTZ, RpcParamDataTypeTimestampAlias, RpcParamDataTypeTimestampTZAlias:
		return "time.Time"
	case RpcParamDataTypeJSON, RpcParamDataTypeJSONB:
		return "map[string]interface{}"
	// ARRAY TYPES
	case RpcParamDataTypeArrayInteger, RpcParamDataTypeArrayBigInt:
		return "[]int64"
	case RpcParamDataTypeArrayReal:
		return "[]float32"
	case RpcParamDataTypeArrayDoublePreci, RpcParamDataTypeArrayNumeric:
		return "[]float64"
	case RpcParamDataTypeArrayText, RpcParamDataTypeArrayVarchar, RpcParamDataTypeArrayVarcharAlias:
		return "[]string"
	case RpcParamDataTypeArrayBoolean:
		return "[]bool"
	case RpcParamDataTypeArrayBytea:
		return "[][]byte"
	case RpcParamDataTypeArrayTimestamp, RpcParamDataTypeArrayTimestampTZ, RpcParamDataTypeArrayTimestampAlias, RpcParamDataTypeArrayTimestampTZAlias:
		return "[]time.Time"
	case RpcParamDataTypeArrayJSON, RpcParamDataTypeArrayJSONB:
		return "[]map[string]interface{}"
	default:
		return "interface{}" // Return interface{} for unknown types
	}
}

func GetValidRpcParamType(pType string, returnAlias bool) (RpcParamDataType, error) {
	pCheckType := RpcParamDataType(strings.ToUpper(pType))
	switch pCheckType {
	case RpcParamDataTypeInteger:
		return RpcParamDataTypeInteger, nil
	case RpcParamDataTypeBigInt:
		return RpcParamDataTypeBigInt, nil
	case RpcParamDataTypeReal:
		return RpcParamDataTypeReal, nil
	case RpcParamDataTypeDoublePreci:
		return RpcParamDataTypeDoublePreci, nil
	case RpcParamDataTypeNumeric:
		return RpcParamDataTypeNumeric, nil
	case RpcParamDataTypeText:
		return RpcParamDataTypeText, nil
	case RpcParamDataTypeVarchar, RpcParamDataTypeVarcharAlias:
		if returnAlias {
			return RpcParamDataTypeVarcharAlias, nil
		}
		return RpcParamDataTypeVarchar, nil
	case RpcParamDataTypeBoolean:
		return RpcParamDataTypeBoolean, nil
	case RpcParamDataTypeBytea:
		return RpcParamDataTypeBytea, nil
	case RpcParamDataTypeTimestamp, RpcParamDataTypeTimestampAlias:
		if returnAlias {
			return RpcParamDataTypeTimestampAlias, nil
		}
		return RpcParamDataTypeTimestamp, nil
	case RpcParamDataTypeTimestampTZ, RpcParamDataTypeTimestampTZAlias:
		if returnAlias {
			return RpcParamDataTypeTimestampTZAlias, nil
		}
		return RpcParamDataTypeTimestampTZ, nil
	case RpcParamDataTypeJSON:
		return RpcParamDataTypeJSON, nil
	case RpcParamDataTypeJSONB:
		return RpcParamDataTypeJSONB, nil
	// ARRAY TYPES
	case RpcParamDataTypeArrayInteger:
		return RpcParamDataTypeArrayInteger, nil
	case RpcParamDataTypeArrayBigInt:
		return RpcParamDataTypeArrayBigInt, nil
	case RpcParamDataTypeArrayReal:
		return RpcParamDataTypeArrayReal, nil
	case RpcParamDataTypeArrayDoublePreci:
		return RpcParamDataTypeArrayDoublePreci, nil
	case RpcParamDataTypeArrayNumeric:
		return RpcParamDataTypeArrayNumeric, nil
	case RpcParamDataTypeArrayText:
		return RpcParamDataTypeArrayText, nil
	case RpcParamDataTypeArrayVarchar, RpcParamDataTypeArrayVarcharAlias:
		if returnAlias {
			return RpcParamDataTypeArrayVarcharAlias, nil
		}
		return RpcParamDataTypeArrayVarchar, nil
	case RpcParamDataTypeArrayBoolean:
		return RpcParamDataTypeArrayBoolean, nil
	case RpcParamDataTypeArrayBytea:
		return RpcParamDataTypeArrayBytea, nil
	case RpcParamDataTypeArrayTimestamp, RpcParamDataTypeArrayTimestampAlias:
		if returnAlias {
			return RpcParamDataTypeArrayTimestampAlias, nil
		}
		return RpcParamDataTypeArrayTimestamp, nil
	case RpcParamDataTypeArrayTimestampTZ, RpcParamDataTypeArrayTimestampTZAlias:
		if returnAlias {
			return RpcParamDataTypeArrayTimestampTZAlias, nil
		}
		return RpcParamDataTypeArrayTimestampTZ, nil
	case RpcParamDataTypeArrayJSON:
		return RpcParamDataTypeArrayJSON, nil
	case RpcParamDataTypeArrayJSONB:
		return RpcParamDataTypeArrayJSONB, nil
	default:
		return "", fmt.Errorf("unsupported rpc param type  : %s", pCheckType)
	}
}

func RpcReturnToGoType(dataType RpcReturnDataType) string {
	switch dataType {
	case RpcReturnDataTypeInteger, RpcReturnDataTypeBigInt:
		return "int64"
	case RpcReturnDataTypeReal:
		return "float32"
	case RpcReturnDataTypeDoublePreci, RpcReturnDataTypeNumeric:
		return "float64"
	case RpcReturnDataTypeText, RpcReturnDataTypeVarchar:
		return "string"
	case RpcReturnDataTypeBoolean:
		return "bool"
	case RpcReturnDataTypeBytea:
		return "[]byte"
	case RpcReturnDataTypeTimestamp, RpcReturnDataTypeTimestampTZ:
		return "time.Time"
	case RpcReturnDataTypeJSON, RpcReturnDataTypeJSONB:
		return "map[string]interface{}"
	// ARRAY TYPES
	case RpcReturnDataTypeArrayInteger, RpcReturnDataTypeArrayBigInt:
		return "[]int64"
	case RpcReturnDataTypeArrayReal:
		return "[]float32"
	case RpcReturnDataTypeArrayDoublePreci, RpcReturnDataTypeArrayNumeric:
		return "[]float64"
	case RpcReturnDataTypeArrayText, RpcReturnDataTypeArrayVarchar:
		return "[]string"
	case RpcReturnDataTypeArrayBoolean:
		return "[]bool"
	case RpcReturnDataTypeArrayBytea:
		return "[][]byte"
	case RpcReturnDataTypeArrayTimestamp, RpcReturnDataTypeArrayTimestampTZ:
		return "[]time.Time"
	case RpcReturnDataTypeArrayJSON, RpcReturnDataTypeArrayJSONB:
		return "[]map[string]interface{}"
	default:
		return "interface{}" // Return interface{} for unknown types
	}
}

func GetValidRpcReturnType(pType string, returnAlias bool) (RpcReturnDataType, error) {
	pCheckType := RpcReturnDataType(strings.ToUpper(pType))
	switch pCheckType {
	case RpcReturnDataTypeInteger:
		return RpcReturnDataTypeInteger, nil
	case RpcReturnDataTypeBigInt:
		return RpcReturnDataTypeBigInt, nil
	case RpcReturnDataTypeReal:
		return RpcReturnDataTypeReal, nil
	case RpcReturnDataTypeDoublePreci:
		return RpcReturnDataTypeDoublePreci, nil
	case RpcReturnDataTypeNumeric:
		return RpcReturnDataTypeNumeric, nil
	case RpcReturnDataTypeText:
		return RpcReturnDataTypeText, nil
	case RpcReturnDataTypeVarchar, RpcReturnDataTypeVarcharAlias:
		if returnAlias {
			return RpcReturnDataTypeVarcharAlias, nil
		}
		return RpcReturnDataTypeVarchar, nil
	case RpcReturnDataTypeBoolean:
		return RpcReturnDataTypeBoolean, nil
	case RpcReturnDataTypeBytea:
		return RpcReturnDataTypeBytea, nil
	case RpcReturnDataTypeTimestamp, RpcReturnDataTypeTimestampAlias:
		if returnAlias {
			return RpcReturnDataTypeTimestampAlias, nil
		}
		return RpcReturnDataTypeTimestamp, nil
	case RpcReturnDataTypeTimestampTZ, RpcReturnDataTypeTimestampTZAlias:
		if returnAlias {
			return RpcReturnDataTypeTimestampTZAlias, nil
		}
		return RpcReturnDataTypeTimestampTZ, nil
	case RpcReturnDataTypeJSON:
		return RpcReturnDataTypeJSON, nil
	case RpcReturnDataTypeJSONB:
		return RpcReturnDataTypeJSONB, nil
	case RpcReturnDataTypeSetOf:
		return RpcReturnDataTypeSetOf, nil
	case RpcReturnDataTypeTable:
		return RpcReturnDataTypeTable, nil
	case RpcReturnDataTypeVoid:
		return RpcReturnDataTypeVoid, nil
		// ARRAY TYPES
	case RpcReturnDataTypeArrayInteger:
		return RpcReturnDataTypeArrayInteger, nil
	case RpcReturnDataTypeArrayBigInt:
		return RpcReturnDataTypeArrayBigInt, nil
	case RpcReturnDataTypeArrayReal:
		return RpcReturnDataTypeArrayReal, nil
	case RpcReturnDataTypeArrayDoublePreci:
		return RpcReturnDataTypeArrayDoublePreci, nil
	case RpcReturnDataTypeArrayNumeric:
		return RpcReturnDataTypeArrayNumeric, nil
	case RpcReturnDataTypeArrayText:
		return RpcReturnDataTypeArrayText, nil
	case RpcReturnDataTypeArrayVarchar, RpcReturnDataTypeArrayVarcharAlias:
		if returnAlias {
			return RpcReturnDataTypeArrayVarcharAlias, nil
		}
		return RpcReturnDataTypeArrayVarchar, nil
	case RpcReturnDataTypeArrayBoolean:
		return RpcReturnDataTypeArrayBoolean, nil
	case RpcReturnDataTypeArrayBytea:
		return RpcReturnDataTypeArrayBytea, nil
	case RpcReturnDataTypeArrayTimestamp, RpcReturnDataTypeArrayTimestampAlias:
		if returnAlias {
			return RpcReturnDataTypeArrayTimestampAlias, nil
		}
		return RpcReturnDataTypeArrayTimestamp, nil
	case RpcReturnDataTypeArrayTimestampTZ, RpcReturnDataTypeArrayTimestampTZAlias:
		if returnAlias {
			return RpcReturnDataTypeArrayTimestampTZAlias, nil
		}
		return RpcReturnDataTypeArrayTimestampTZ, nil
	case RpcReturnDataTypeArrayJSON:
		return RpcReturnDataTypeArrayJSON, nil
	case RpcReturnDataTypeArrayJSONB:
		return RpcReturnDataTypeArrayJSONB, nil
	default:
		return "", fmt.Errorf("unsupported rpc return type  : %s", pCheckType)
	}
}

func GetValidRpcReturnNameDecl(pType RpcReturnDataType, returnAlias bool) (string, error) {
	switch pType {
	case RpcReturnDataTypeInteger:
		return "RpcReturnDataTypeInteger", nil
	case RpcReturnDataTypeBigInt:
		return "RpcReturnDataTypeBigInt", nil
	case RpcReturnDataTypeReal:
		return "RpcReturnDataTypeReal", nil
	case RpcReturnDataTypeDoublePreci:
		return "RpcReturnDataTypeDoublePreci", nil
	case RpcReturnDataTypeNumeric:
		return "RpcReturnDataTypeNumeric", nil
	case RpcReturnDataTypeText:
		return "RpcReturnDataTypeText", nil
	case RpcReturnDataTypeVarchar, RpcReturnDataTypeVarcharAlias:
		if returnAlias {
			return "RpcReturnDataTypeVarcharAlias", nil
		}
		return "RpcReturnDataTypeVarchar", nil
	case RpcReturnDataTypeBoolean:
		return "RpcReturnDataTypeBoolean", nil
	case RpcReturnDataTypeBytea:
		return "RpcReturnDataTypeBytea", nil
	case RpcReturnDataTypeTimestamp:
		if returnAlias {
			return "RpcReturnDataTypeTimestampAlias", nil
		}
		return "RpcReturnDataTypeTimestamp", nil
	case RpcReturnDataTypeTimestampTZ:
		if returnAlias {
			return "RpcReturnDataTypeTimestampTZAlias", nil
		}
		return "RpcReturnDataTypeTimestampTZ", nil
	case RpcReturnDataTypeJSON:
		return "RpcReturnDataTypeJSON", nil
	case RpcReturnDataTypeJSONB:
		return "RpcReturnDataTypeJSONB", nil
	case RpcReturnDataTypeSetOf:
		return "RpcReturnDataTypeSetOf", nil
	case RpcReturnDataTypeTable:
		return "RpcReturnDataTypeTable", nil
	case RpcReturnDataTypeVoid:
		return "RpcReturnDataTypeVoid", nil
	// ARRAY TYPES
	case RpcReturnDataTypeArrayInteger:
		return "RpcReturnDataTypeArrayInteger", nil
	case RpcReturnDataTypeArrayBigInt:
		return "RpcReturnDataTypeArrayBigInt", nil
	case RpcReturnDataTypeArrayReal:
		return "RpcReturnDataTypeArrayReal", nil
	case RpcReturnDataTypeArrayDoublePreci:
		return "RpcReturnDataTypeArrayDoublePreci", nil
	case RpcReturnDataTypeArrayNumeric:
		return "RpcReturnDataTypeArrayNumeric", nil
	case RpcReturnDataTypeArrayText:
		return "RpcReturnDataTypeArrayText", nil
	case RpcReturnDataTypeArrayVarchar, RpcReturnDataTypeArrayVarcharAlias:
		if returnAlias {
			return "RpcReturnDataTypeArrayVarcharAlias", nil
		}
		return "RpcReturnDataTypeArrayVarchar", nil
	case RpcReturnDataTypeArrayBoolean:
		return "RpcReturnDataTypeArrayBoolean", nil
	case RpcReturnDataTypeArrayBytea:
		return "RpcReturnDataTypeArrayBytea", nil
	case RpcReturnDataTypeArrayTimestamp:
		if returnAlias {
			return "RpcReturnDataTypeArrayTimestampAlias", nil
		}
		return "RpcReturnDataTypeArrayTimestamp", nil
	case RpcReturnDataTypeArrayTimestampTZ:
		if returnAlias {
			return "RpcReturnDataTypeArrayTimestampTZAlias", nil
		}
		return "RpcReturnDataTypeArrayTimestampTZ", nil
	case RpcReturnDataTypeArrayJSON:
		return "RpcReturnDataTypeArrayJSON", nil
	case RpcReturnDataTypeArrayJSONB:
		return "RpcReturnDataTypeArrayJSONB", nil
	default:
		return "", fmt.Errorf("unsupported rpc return name declaration  : %s", pType)
	}
}

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

	RpcParams []RpcParam

	RpcModel struct {
		Alias string
		Model any
	}

	Rpc interface {
		BindModels()
		BindModel(model any, alias string) Rpc
		GetModels() map[string]RpcModel
		SetName(name string)
		GetName() string
		SetParams(params []RpcParam)
		GetParams() []RpcParam
		UseParamPrefix() bool
		SetSchema(schema string)
		GetSchema() string
		SetSecurity(security RpcSecurityType)
		GetSecurity() RpcSecurityType
		SetBehavior(behavior RpcBehaviorType)
		GetBehavior() RpcBehaviorType
		SetReturnType(returnType RpcReturnDataType)
		GetReturnType() RpcReturnDataType
		SetReturnTypeStmt(returnTypeStmt string)
		GetReturnTypeStmt() string
		SetRawDefinition(definition string)
		GetRawDefinition() string
		SetCompleteStmt(stmt string)
		GetCompleteStmt() string
	}

	RpcBase struct {
		Name              string
		Schema            string
		Params            []RpcParam
		Definition        string
		SecurityType      RpcSecurityType
		ReturnType        RpcReturnDataType
		ReturnTypeStmt    string
		Behavior          RpcBehaviorType
		CompleteStatement string
		Models            map[string]RpcModel
	}

	RpcParamTag struct {
		Name         string
		Type         string
		DefaultValue string
	}
)

var (
	DefaultRpcParamPrefix = "in_"
	DefaultRpcSchema      = "public"
)

const (
	RpcBehaviorVolatile  RpcBehaviorType = "VOLATILE"
	RpcBehaviorStable    RpcBehaviorType = "STABLE"
	RpcBehaviorImmutable RpcBehaviorType = "IMMUTABLE"

	RpcSecurityTypeDefiner RpcSecurityType = "DEFINER"
	RpcSecurityTypeInvoker RpcSecurityType = "INVOKER"

	RpcTemplate = `CREATE OR REPLACE FUNCTION :function_name(:params) RETURNS :return_type LANGUAGE plpgsql :behavior :security AS $function$ :definition $function$`
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
				return paramTag, err
			}
			paramTag.Type = string(pType)
		case "default":
			paramTag.DefaultValue = value
		}

	}
	return paramTag, nil
}

// ----- Rpc base functionality -----
func (r *RpcBase) initModel() {
	if r.Models == nil {
		r.Models = make(map[string]RpcModel)
	}
}

func (r *RpcBase) SetName(name string) {
	r.Name = name
}

func (r *RpcBase) GetName() string {
	return r.Name
}

func (r *RpcBase) BindModel(model any, alias string) Rpc {
	r.initModel()

	reflectType := reflect.TypeOf(model)
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}

	r.Models[utils.ToSnakeCase(reflectType.Name())] = RpcModel{
		Alias: alias,
		Model: model,
	}

	return r
}

func (r *RpcBase) BindModels() {}
func (r *RpcBase) GetModels() map[string]RpcModel {
	return r.Models
}

func (r *RpcBase) SetReturnTypeStmt(returnTypeStmt string) {
	r.ReturnTypeStmt = returnTypeStmt
}

func (r *RpcBase) GetReturnTypeStmt() string {
	return r.ReturnTypeStmt
}

func (r *RpcBase) SetParams(params []RpcParam) {
	r.Params = append(r.Params, params...)
}
func (r *RpcBase) GetParams() []RpcParam {
	return r.Params
}

func (r *RpcBase) UseParamPrefix() bool {
	return true
}

func (r *RpcBase) GetReturnType() (rt RpcReturnDataType) {
	RpcLogger.Error("Rpc return type is not implemented, use GetReturnType for set it")
	return
}

func (r *RpcBase) SetSchema(schema string) {
	r.Schema = schema
}

func (r *RpcBase) GetSchema() string {
	return r.Schema
}

func (r *RpcBase) SetSecurity(security RpcSecurityType) {
	r.SecurityType = security
}

func (r *RpcBase) GetSecurity() RpcSecurityType {
	return r.SecurityType
}

func (r *RpcBase) SetBehavior(behavior RpcBehaviorType) {
	r.Behavior = behavior
}

func (r *RpcBase) GetBehavior() RpcBehaviorType {
	return RpcBehaviorVolatile
}

func (r *RpcBase) SetReturnType(returnType RpcReturnDataType) {
	r.ReturnType = returnType
}

func (r *RpcBase) SetRawDefinition(definition string) {
	r.Definition = definition
}

func (r *RpcBase) GetRawDefinition() (d string) {
	RpcLogger.Error("Rpc definition type is not implemented, use GetRawDefinition for set it")
	return
}

func (r *RpcBase) SetCompleteStmt(stmt string) {
	r.CompleteStatement = stmt
}

func (r *RpcBase) GetCompleteStmt() string {
	return r.CompleteStatement
}

// ----- Rpc Param Functionality -----
func (p RpcParams) ToQuery(userPrefix bool) (string, error) {
	var qArr []string
	for i := range p {
		pi := p[i]

		var prefix string
		if userPrefix {
			prefix = DefaultRpcParamPrefix
		}

		pt, err := GetValidRpcParamType(string(pi.Type), false)
		if err != nil {
			return "", err
		}

		pStr := fmt.Sprintf("%s%s %s", prefix, pi.Name, pt)
		if pi.Default != nil {
			pStr += fmt.Sprintf(" default '%s'::%s", *pi.Default, string(pt))
		}

		qArr = append(qArr, pStr)
	}

	return strings.Join(qArr, ", "), nil
}

func BuildRpc(rpc Rpc) (err error) {
	rpc.BindModels()

	// init value from template
	q := RpcTemplate

	// set rpc type and value
	rpcType := reflect.TypeOf(rpc)
	if rpcType.Kind() == reflect.Ptr {
		rpcType = rpcType.Elem()
	}

	// set rpc name
	rpcName := rpc.GetName()
	if rpcName == "" {
		rpcName = utils.ToSnakeCase(rpcType.Name())
	}
	rpc.SetName(rpcName)

	// replace enhance rpcName and set rpc base schema
	if rpc.GetSchema() == "" {
		rpc.SetSchema(DefaultRpcSchema)
	}
	rpcName = fmt.Sprintf("%s.%s", rpc.GetSchema(), rpcName)

	// replace definition and set rpc base name
	q = strings.ReplaceAll(q, ":function_name", rpcName)

	// build Param
	pt, found := rpcType.FieldByName("Params")
	if !found {
		return fmt.Errorf("field Params is not found in struct : %s", rpcType.Name())
	}

	if p, err := extractRpcParam(pt.Type); err != nil {
		return err
	} else {
		pq, ep := p.ToQuery(rpc.UseParamPrefix())
		if ep != nil {
			return ep
		}

		// replace param definition and set rpc base param
		rpc.SetParams(p)
		q = strings.ReplaceAll(q, ":params", strings.ToLower(pq))
	}

	// build return data
	rt, found := rpcType.FieldByName("Return")
	if !found {
		return fmt.Errorf("field Return is not found in struct : %s", rpcType.Name())
	}

	if rType, err := extractRpcResult(rt.Type, rpc); err != nil {
		return err
	} else {
		// replace return type definition and set rpc base return type
		rpc.SetReturnType(rpc.GetReturnType())
		rpc.SetReturnTypeStmt(strings.ToLower(rType))
		q = strings.ReplaceAll(q, ":return_type", strings.ToLower(rType))
	}

	// build security
	if rpc.GetSecurity() == "" {
		rpc.SetSecurity(RpcSecurityTypeInvoker)
	}

	if rpc.GetSecurity() == RpcSecurityTypeDefiner {
		rpc.SetSecurity(RpcSecurityTypeDefiner)
		q = strings.ReplaceAll(q, ":security", "SECURITY DEFINER")
	} else {
		q = strings.ReplaceAll(q, ":security", "")
	}

	// set behavior
	if rpc.GetBehavior() == "" {
		rpc.SetBehavior(RpcBehaviorVolatile)
	} else {
		rpc.SetBehavior(rpc.GetBehavior())
	}
	if rpc.GetBehavior() == RpcBehaviorVolatile {
		q = strings.ReplaceAll(q, ":behavior", "")
	} else {
		q = strings.ReplaceAll(q, ":behavior", string(rpc.GetBehavior()))
	}

	// build definitions
	definition := buildRpcDefinition(rpc)
	rpc.SetRawDefinition(definition)
	q = strings.ReplaceAll(q, ":definition", definition)

	// cleanup
	re := regexp.MustCompile(`\s+`)
	q = re.ReplaceAllString(q, " ")
	q = strings.ToLower(q)
	rpc.SetCompleteStmt(q)

	return
}

func extractRpcParam(paramType reflect.Type) (params RpcParams, err error) {
	if paramType.Kind() == reflect.Pointer {
		paramType = paramType.Elem()
	}

	for i := 0; i < paramType.NumField(); i++ {
		field := paramType.Field(i)

		columnTagStr := field.Tag.Get("column")
		if len(columnTagStr) == 0 {
			continue
		}

		ct, err := UnmarshalRpcParamTag(columnTagStr)
		if err != nil {
			return params, err
		}

		p := RpcParam{
			Name: ct.Name,
			Type: RpcParamDataType(ct.Type),
		}

		if ct.DefaultValue != "" {
			p.Default = &ct.DefaultValue
		}

		params = append(params, p)
	}

	return
}

func extractRpcResult(returnReflectType reflect.Type, rpc Rpc) (q string, err error) {
	switch rpc.GetReturnType() {
	case RpcReturnDataTypeSetOf:
		return buildRpcReturnSetOf(returnReflectType)
	case RpcReturnDataTypeTable:
		return buildRpcReturnTable(returnReflectType)
	default:
		return string(rpc.GetReturnType()), nil
	}
}

func buildRpcReturnSetOf(returnReflectType reflect.Type) (q string, err error) {
	st, err := findStruct(returnReflectType)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("setof %s", st.Name()), nil
}

func buildRpcReturnTable(returnReflectType reflect.Type) (q string, err error) {
	st, e := findStruct(returnReflectType)
	if e != nil {
		err = e
		return
	}

	p, e := extractRpcParam(st)
	if e != nil {
		err = e
		return
	}

	pq, ep := p.ToQuery(false)
	if ep != nil {
		return q, ep
	}

	if len(pq) == 0 {
		return
	}

	return fmt.Sprintf("table(%s)", pq), nil
}

func buildRpcDefinition(rpc Rpc) string {
	definition := rpc.GetRawDefinition()

	dFields := strings.Fields(utils.CleanUpString(definition))
	for i := range dFields {
		d := dFields[i]
		if strings.HasSuffix(d, ";") && strings.ToLower(d) != "end;" {
			dFields[i] = strings.ReplaceAll(d, ";", " ;")
		}
	}
	definition = strings.Join(dFields, " ")

	for k, v := range rpc.GetModels() {
		definition = utils.MatchReplacer(definition, ":"+v.Alias, k)
	}

	params := rpc.GetParams()
	for i := range params {
		p := params[i]
		key := p.Name
		replaceKey := key
		if rpc.UseParamPrefix() {
			replaceKey = DefaultRpcParamPrefix + key
		}

		definition = utils.MatchReplacer(definition, ":"+key, replaceKey)
	}

	return definition
}

func findStruct(returnReflectType reflect.Type) (reflect.Type, error) {
	switch returnReflectType.Kind() {
	case reflect.Ptr:
		return findStruct(returnReflectType.Elem())
	case reflect.Array, reflect.Slice:
		return findStruct(returnReflectType.Elem())
	case reflect.Struct:
		return returnReflectType, nil
	default:
		return nil, fmt.Errorf("%s is not struct", returnReflectType.Name())
	}
}

// ----- Execute Rpc -----
func ExecuteRpc(ctx Context, rpc Rpc) (any, error) {
	rpcType := reflect.TypeOf(rpc).Elem()
	rpcValue := reflect.ValueOf(rpc).Elem()
	if rpcType.Kind() == reflect.Pointer {
		rpcType = rpcType.Elem()
		rpcValue = rpcValue.Elem()
	}

	// set params
	paramsFields, found := rpcType.FieldByName("Params")
	if !found {
		return nil, &ErrorResponse{
			StatusCode: fasthttp.StatusInternalServerError,
			Details:    fmt.Sprintf("Struct %s doesn`t have Params field, define first because this attribute need for send parameter to server", rpcType.Name()),
			Message:    fmt.Sprintf("Undefined field Params in struct %s", rpcType.Name()),
			Hint:       "Invalid Rpc",
			Code:       fasthttp.StatusMessage(fasthttp.StatusInternalServerError),
		}
	}
	paramsType := paramsFields.Type
	paramValue := rpcValue.FieldByName("Params")
	if paramsType.Kind() == reflect.Ptr {
		paramsType = paramsType.Elem()
		paramValue = paramValue.Elem()
	}

	returnField, found := rpcType.FieldByName("Return")
	if !found {
		return nil, &ErrorResponse{
			StatusCode: fasthttp.StatusInternalServerError,
			Details:    fmt.Sprintf("Struct %s doesn`t have Return field, define first because this attribute need for receive data from server", rpcType.Name()),
			Message:    fmt.Sprintf("Undefined field Return in struct %s", rpcType.Name()),
			Hint:       "Invalid Rpc",
			Code:       fasthttp.StatusMessage(fasthttp.StatusInternalServerError),
		}
	}

	if err := BuildRpc(rpc); err != nil {
		return nil, err
	}

	mapParams := map[string]any{}
	for i := 0; i < paramsType.NumField(); i++ {
		if paramValue.IsValid() {
			fieldType, fieldValue := paramsType.Field(i), paramValue.Field(i)
			key := utils.SnakeCaseToPascalCase(fieldType.Name)
			if rpc.UseParamPrefix() {
				key = fmt.Sprintf("%s%s", DefaultRpcParamPrefix, key)
			}
			mapParams[strings.ToLower(key)] = fieldValue.Interface()
		}
	}

	pByte, err := json.Marshal(mapParams)
	if err != nil {
		return nil, &ErrorResponse{
			StatusCode: fasthttp.StatusBadRequest,
			Details:    err.Error(),
			Message:    "Invalid request data",
			Hint:       "Invalid params",
			Code:       fasthttp.StatusMessage(fasthttp.StatusBadRequest),
		}
	}

	apiUrl := fmt.Sprintf("%s/%s/%s", ctx.Config().SupabasePublicUrl, "rest/v1/rpc", rpc.GetName())
	if string(ctx.RequestContext().QueryArgs().QueryString()) != "" {
		apiUrl = fmt.Sprintf("%s?%s", apiUrl, string(ctx.RequestContext().QueryArgs().QueryString()))
	}

	httpReq, err := ConvertRequestCtxToHTTPRequest(ctx.RequestContext())
	if err != nil {
		return nil, err
	}

	resData, err := rpcSendRequest(apiUrl, pByte, rpcAttachAuthHeader(httpReq))
	if err != nil {
		return nil, err
	}

	// sample data
	returnObject := reflect.New(returnField.Type).Interface()
	if err := json.Unmarshal(resData, returnObject); err != nil {
		return nil, &ErrorResponse{
			StatusCode: fasthttp.StatusInternalServerError,
			Details:    err,
			Message:    "invalid marshall response data",
		}
	}

	returnValue := reflect.ValueOf(returnObject)
	if returnValue.Kind() == reflect.Ptr {
		returnValue = returnValue.Elem()
	}

	rv := rpcValue.FieldByName("Return")
	rv.Set(returnValue)

	return returnValue.Interface(), nil
}

func rpcAttachAuthHeader(inReq *http.Request) net.RequestInterceptor {
	return func(outReq *http.Request) error {
		if authHeader := inReq.Header.Get("Authorization"); len(authHeader) > 0 {
			outReq.Header.Set("Authorization", authHeader)
		}

		if apiKey := inReq.Header.Get("apiKey"); len(apiKey) > 0 {
			outReq.Header.Set("apiKey", apiKey)
		}

		return nil
	}
}

func rpcSendRequest(apiUrl string, body []byte, reqInterceptor net.RequestInterceptor) ([]byte, error) {
	resData, err := net.SendRequest(fasthttp.MethodPost, apiUrl, body, net.DefaultTimeout, reqInterceptor, nil)
	if err != nil {
		sendErr, isHaveData := err.(utils.SendRequestError)
		if isHaveData {
			var errResponse ErrorResponse
			if err := json.Unmarshal(sendErr.Body, &errResponse); err == nil {
				return nil, &errResponse
			}
		}

		return nil, &ErrorResponse{
			StatusCode: fasthttp.StatusInternalServerError,
			Details:    err.Error(),
			Message:    fmt.Sprintf("fail request to upstream. Reason: %v", err),
		}
	}
	return resData, nil
}

func ConvertRequestCtxToHTTPRequest(ctx *fasthttp.RequestCtx) (*http.Request, error) {
	url, err := url.ParseRequestURI(string(ctx.RequestURI()))
	if err != nil {
		return nil, err
	}
	// Create a new http.Request based on the data in RequestCtx
	req := &http.Request{
		Method:     string(ctx.Method()),
		URL:        url,
		Proto:      "HTTP/1.1", // You may need to adjust this based on your requirements
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	// Copy headers from RequestCtx to http.Request
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		req.Header.Add(string(key), string(value))
	})

	// Copy body from RequestCtx to http.Request
	req.Body = io.NopCloser(bytes.NewReader(ctx.Request.Body()))

	return req, nil
}
