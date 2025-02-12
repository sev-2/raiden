package postgres

import (
	"fmt"
	"strings"
)

type DataType string

const (
	// ----- Number Type -----

	// smallIntType represents a small-range integer in PostgreSQL.
	SmallIntType DataType = "smallint" // 2 bytes, range: -32768 to +32767

	// intType represents a typical choice for an integer in PostgreSQL.
	IntType DataType = "integer" // 4 bytes, range: -2147483648 to +2147483647

	// bigIntType represents a large-range integer in PostgreSQL.
	BigIntType DataType = "bigint" // 8 bytes, range: -9223372036854775808 to +9223372036854775807

	// decimalType represents a user-specified precision decimal in PostgreSQL.
	DecimalType DataType = "decimal" // Variable precision, exact, up to 131072 digits before the decimal point; up to 16383 digits after the decimal point

	// numericType represents a user-specified precision numeric in PostgreSQL.
	NumericType DataType = "numeric" // Variable precision, exact, up to 131072 digits before the decimal point; up to 16383 digits after the decimal point

	// realType represents a variable-precision, inexact real number in PostgreSQL.
	RealType DataType = "real" // 4 bytes, 6 decimal digits precision

	// doublePrecisionType represents a variable-precision, inexact double precision number in PostgreSQL.
	DoublePrecisionType      DataType = "double precision" // 8 bytes, 15 decimal digits precision
	DoublePrecisionTypeAlias DataType = "float8"

	// smallSerialType represents a small auto-incrementing integer in PostgreSQL.
	SmallSerialType DataType = "smallserial" // 2 bytes, range: 1 to 32767

	// serialType represents an auto-incrementing integer in PostgreSQL.
	SerialType DataType = "serial" // 4 bytes, range: 1 to 2147483647

	// bigSerialType represents a large auto-incrementing integer in PostgreSQL.
	BigSerialType DataType = "bigserial" // 8 bytes, range: 1 to 9223372036854775807

	// ----- Character Type -----

	// varcharType represents a variable-length character type with a specified limit in PostgreSQL.
	VarcharType      DataType = "character varying" // Variable-length with limit
	VarcharTypeAlias DataType = "varchar"

	// charType represents a fixed-length character type in PostgreSQL.
	CharType DataType = "char" // Fixed-length, blank-padded

	// bpcharType represents a fixed-length character type (blank-padded) in PostgreSQL.
	BpcharType DataType = "bpchar" // Fixed-length, blank-padded (alternative representation)

	// textType represents a variable-length character type with unlimited length in PostgreSQL.
	TextType DataType = "text" // Variable unlimited length

	// ----- Time Type -----

	// timestampType represents a timestamp with both date and time in PostgreSQL.
	TimestampType      DataType = "timestamp without time zone" // 8 bytes, date and time (no time zone), 1 microsecond
	TimestampTypeAlias DataType = "timestamp"

	// timestampTzType represents a timestamp with both date and time, with time zone in PostgreSQL.
	TimestampTzType      DataType = "timestamp with time zone" // 8 bytes, date and time with time zone, 1 microsecond
	TimestampTzTypeAlias DataType = "timestampz"

	// dateType represents a date (no time of day) in PostgreSQL.
	DateType DataType = "date" // 4 bytes, date (no time of day), 1 day

	// timeType represents a time of day (no date) in PostgreSQL.
	TimeType      DataType = "time without time zone" // 8 bytes, time of day (no date), 1 microsecond
	TimeTypeAlias DataType = "time"
	// timeTzType represents a time of day (no date), with time zone in PostgreSQL.
	TimeTzType      DataType = "time with time zone" // 12 bytes, time of day with time zone, 1 microsecond
	TimeTzTypeAlias DataType = "timez"

	// intervalType represents a time interval in PostgreSQL.
	IntervalType DataType = "interval" // 16 bytes, time interval, 1 microsecond

	// ----- Boolean Type -----
	BooleanType DataType = "boolean"

	// ----- Uuid Type -----

	// uuidType represents the UUID data type in PostgreSQL.
	UuidType DataType = "uuid"

	// ----- Json Type -----

	// jsonType represents the JSON data type in PostgreSQL.
	JsonType DataType = "json"

	// jsonbType represents the JSONB data type in PostgreSQL.
	JsonbType DataType = "jsonb"

	PointType DataType = "point"

	UserDefined DataType = "USER-DEFINED"
)

// ToGoType Convert postgres type to golang type
func ToGoType(pgType DataType, isNullable bool) (goType string) {
	switch pgType {
	case SmallIntType, SerialType, SmallSerialType:
		goType = "int16"
	case IntType:
		goType = "int32"
	case BigIntType, BigSerialType:
		goType = "int64"
	case DecimalType, NumericType, RealType, DoublePrecisionType:
		goType = "float64"
	case VarcharType, VarcharTypeAlias, CharType, BpcharType, TextType:
		goType = "string"
	case TimestampType, TimestampTypeAlias, TimestampTzType, TimestampTzTypeAlias, TimeType, TimeTypeAlias, TimeTzType, TimeTzTypeAlias, DateType:
		goType = "time.Time"
	case IntervalType:
		goType = "time.Duration"
	case BooleanType:
		goType = "bool"
	case UuidType:
		goType = "uuid.UUID" // Assuming you have a UUID library imported
	case JsonType, JsonbType:
		goType = "interface{}" // Use a more specific type based on your JSON library
	case PointType:
		goType = "postgres.Point"
	default:
		goType = "interface{}"
	}

	if isNullable && goType != "interface{}" {
		goType = fmt.Sprintf("*%s", goType)
	}

	return
}

// ToPostgresType converts a Go type to its corresponding PostgreSQL data type.
func ToPostgresType(goType string) (pgType DataType) {
	switch goType {
	case "int16":
		pgType = SmallIntType
	case "int32":
		pgType = IntType
	case "int64":
		pgType = BigIntType
	case "uint16":
		pgType = SmallSerialType
	case "uint32":
		pgType = SerialType
	case "uint64":
		pgType = BigSerialType
	case "float32":
		pgType = RealType
	case "float64":
		pgType = DoublePrecisionType
	case "string":
		pgType = TextType
	case "time.Time":
		pgType = TimestampTzType
	case "time.Duration":
		pgType = IntervalType
	case "bool":
		pgType = BooleanType
	case "uuid.UUID":
		pgType = UuidType
	case "postgres.Point":
		pgType = PointType
	case "interface{}", "any":
		pgType = TextType

	default:
		// Default to TEXT for unknown types
		pgType = TextType
	}

	return
}

// IsValidDataType checks if the given value is a valid DataType constant.
func IsValidDataType(value string) bool {
	validDataTypes := map[DataType]struct{}{
		// ----- Number Type -----
		SmallIntType: {}, IntType: {}, BigIntType: {},
		DecimalType: {}, NumericType: {}, RealType: {}, DoublePrecisionType: {}, DoublePrecisionTypeAlias: {},
		SmallSerialType: {}, SerialType: {}, BigSerialType: {},
		// ----- Character Type -----
		VarcharType: {}, VarcharTypeAlias: {}, CharType: {}, BpcharType: {}, TextType: {},
		// ----- Time Type -----
		TimestampType: {}, TimestampTypeAlias: {}, TimestampTzType: {}, TimestampTzTypeAlias: {}, DateType: {},
		TimeType: {}, TimeTypeAlias: {}, TimeTzType: {}, TimeTzTypeAlias: {}, IntervalType: {},
		// ----- Boolean Type -----
		BooleanType: {},
		// ----- Uuid Type -----
		UuidType: {},
		// ----- Json Type -----
		JsonType: {}, JsonbType: {},
		// ----- Point Type -----
		PointType: {},
	}

	dataType := DataType(strings.ToLower(value))
	_, isValid := validDataTypes[dataType]
	return isValid
}

func GetPgDataTypeName(pgType DataType, returnAlias bool) DataType {
	switch pgType {
	case SmallIntType:
		return SmallIntType
	case SerialType:
		return SerialType
	case SmallSerialType:
		return SmallSerialType
	case IntType:
		return IntType
	case BigIntType:
		return BigIntType
	case BigSerialType:
		return BigSerialType
	case DecimalType:
		return DecimalType
	case NumericType:
		return NumericType
	case RealType:
		return RealType
	case DoublePrecisionType, DoublePrecisionTypeAlias:
		if returnAlias {
			return DoublePrecisionTypeAlias
		}
		return DoublePrecisionType
	case VarcharType, VarcharTypeAlias:
		if returnAlias {
			return VarcharTypeAlias
		}
		return VarcharType
	case CharType:
		return CharType
	case BpcharType:
		return BpcharType
	case TextType:
		return TextType
	case TimestampType, TimestampTypeAlias:
		if returnAlias {
			return TimestampTypeAlias
		}
		return TimestampType
	case TimestampTzType, TimestampTzTypeAlias:
		if returnAlias {
			return TimestampTzTypeAlias
		}
		return TimestampTzType
	case TimeType, TimeTypeAlias:
		if returnAlias {
			return TimeTypeAlias
		}
		return TimeType
	case TimeTzType, TimeTzTypeAlias:
		if returnAlias {
			return TimeTzTypeAlias
		}
		return TimeTzType
	case DateType:
		return DateType
	case IntervalType:
		return IntervalType
	case BooleanType:
		return BooleanType
	case UuidType:
		return UuidType
	case JsonType:
		return JsonType
	case JsonbType:
		return JsonbType
	case PointType:
		return PointType
	}

	return TextType
}
