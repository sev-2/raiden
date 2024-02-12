package raiden

import (
	"fmt"
	"strings"
)

type RpcParamDataType string
type RpcReturnDataType string

// Define constants for rpc input data type
const (
	RpcParamDataTypeInteger          RpcParamDataType = "INTEGER"
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
)

// Define constants for rpc return data type
const (
	RpcReturnDataTypeInteger          RpcReturnDataType = "INTEGER"
	RpcReturnDataTypeBigInt           RpcReturnDataType = "BIGINT"
	RpcReturnDataTypeReal             RpcReturnDataType = "REAL"
	RpcReturnDataTypeDoublePreci      RpcReturnDataType = "DOUBLE PRECISION"
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
)

func RpcParamToGoType(dataType RpcParamDataType) string {
	switch dataType {
	case RpcParamDataTypeInteger, RpcParamDataTypeBigInt:
		return "int64"
	case RpcParamDataTypeReal:
		return "float32"
	case RpcParamDataTypeDoublePreci:
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
	case RpcReturnDataTypeDoublePreci:
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
	default:
		return "", fmt.Errorf("unsupported rpc return type  : %s", pType)
	}
}
