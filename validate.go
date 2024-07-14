package raiden

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/valyala/fasthttp"
)

// custom struct for validation function
type ValidatorFunc struct {
	Name      string
	Validator validator.Func
}

// custom type for custom validation function
type WithValidator func(name string, validateFn validator.Func) ValidatorFunc

// validate payload
func Validate(payload any, requestValidators ...ValidatorFunc) error {
	validatorInstance := validator.New()
	if len(requestValidators) > 0 {
		for _, rv := range requestValidators {
			err := validatorInstance.RegisterValidation(rv.Name, rv.Validator)
			if err != nil {
				return err
			}
		}
	}

	if err := validatorInstance.Struct(payload); err != nil {
		validationError, isValid := err.(validator.ValidationErrors)
		if !isValid {
			return err
		}

		mapErrMessage := make(map[string][]string)
		for _, err := range validationError {
			errMessage := getInvalidMessage(err.Field(), err.Tag(), err.Param())
			errors, isExist := mapErrMessage[err.Field()]
			if isExist {
				errors = append(errors, errMessage)
				mapErrMessage[err.Field()] = errors
				continue
			}
			mapErrMessage[err.Field()] = []string{errMessage}
		}

		errKeys := make([]string, 0, len(mapErrMessage))
		for key := range mapErrMessage {
			errKeys = append(errKeys, key)
		}

		errByte, err := json.Marshal(mapErrMessage)
		if err != nil {
			return err
		}

		err = &ErrorResponse{
			StatusCode: fasthttp.StatusBadRequest,
			Code:       "Validation Fail",
			Details:    string(errByte),
			Message:    "invalid payload for key : " + strings.Join(errKeys, ","),
		}

		return err
	}

	return nil
}

// getInvalidMessage is manual mapping for error message base on validation result
// todo : integrate with i18n
func getInvalidMessage(field, tag, param string) (errMessage string) {
	field = strings.ToLower(field)
	switch tag {
	case "required":
		errMessage = fmt.Sprintf("%s is required", field)
	case "min":
		errMessage = fmt.Sprintf("%s should be at least %s", field, param)
	case "max":
		errMessage = fmt.Sprintf("%s should not exceed %s", field, param)
	case "eq":
		errMessage = fmt.Sprintf("%s should be equal to %s", field, param)
	case "eq_ignore_case":
		errMessage = fmt.Sprintf("%s should be equal to the specified value (case-insensitive)", field)
	case "gt":
		errMessage = fmt.Sprintf("%s should be greater than %s", field, param)
	case "gte":
		errMessage = fmt.Sprintf("%s should be greater than or equal %s", field, param)
	case "lt":
		errMessage = fmt.Sprintf("%s should be less than %s", field, param)
	case "lte":
		errMessage = fmt.Sprintf("%s should be less than or equal %s", field, param)
	case "ne":
		errMessage = fmt.Sprintf("%s should not be equal %s", field, param)
	case "ne_ignore_case":
		errMessage = fmt.Sprintf("%s should not be equal to the specified value (case-insensitive)", field)
	case "alpha":
		errMessage = fmt.Sprintf("%s should contain only alphabetical characters", field)
	case "alphanum":
		errMessage = fmt.Sprintf("%s should contain only alphanumeric characters", field)
	case "numeric":
		errMessage = fmt.Sprintf("%s should be a numeric value", field)
	case "boolean":
		errMessage = fmt.Sprintf("%s should be a boolean value (true or false)", field)
	case "alphaunicode":
		errMessage = fmt.Sprintf("%s should contain only alphabetical Unicode characters", field)
	case "alphanumunicode":
		errMessage = fmt.Sprintf("%s should contain only alphanumeric Unicode characters", field)
	case "ascii":
		errMessage = fmt.Sprintf("%s should contain only ASCII characters", field)
	case "contains":
		errMessage = fmt.Sprintf("%s should contain the specified substring", field)
	case "containsany":
		errMessage = fmt.Sprintf("%s should contain any of the specified substrings", field)
	case "containsrune":
		errMessage = fmt.Sprintf("%s should contain the specified Unicode rune", field)
	case "endsnotwith":
		errMessage = fmt.Sprintf("%s should not end with the specified suffix", field)
	case "endswith":
		errMessage = fmt.Sprintf("%s should end with the specified suffix", field)
	case "excludes":
		errMessage = fmt.Sprintf("%s should not contain the specified substring", field)
	case "excludesall":
		errMessage = fmt.Sprintf("%s should not contain any of the specified substrings", field)
	case "excludesrune":
		errMessage = fmt.Sprintf("%s should not contain the specified Unicode rune", field)
	case "lowercase":
		errMessage = fmt.Sprintf("%s should be in lowercase", field)
	case "multibyte":
		errMessage = fmt.Sprintf("%s should contain only multi-byte characters", field)
	case "printascii":
		errMessage = fmt.Sprintf("%s should contain only printable ASCII characters", field)
	case "startsnotwith":
		errMessage = fmt.Sprintf("%s should not start with the specified prefix", field)
	case "startswith":
		errMessage = fmt.Sprintf("%s should start with the specified prefix", field)
	case "uppercase":
		errMessage = fmt.Sprintf("%s should be in uppercase", field)
	case "fqdn":
		errMessage = fmt.Sprintf("%s should be a Full Qualified Domain Name (FQDN)", field)
	case "hostname":
		errMessage = fmt.Sprintf("%s should be a Hostname (RFC 952)", field)
	case "ip":
		errMessage = fmt.Sprintf("%s should be an Internet Protocol Address (IP)", field)
	case "ipv4":
		errMessage = fmt.Sprintf("%s should be an Internet Protocol Address (IPv4)", field)
	case "ipv6":
		errMessage = fmt.Sprintf("%s should be an Internet Protocol Address (IPv6)", field)
	case "mac":
		errMessage = fmt.Sprintf("%s should be a Media Access Control Address (MAC)", field)
	case "uri":
		errMessage = fmt.Sprintf("%s should be a URI String", field)
	case "url":
		errMessage = fmt.Sprintf("%s should be a URL String", field)
	case "base64":
		errMessage = fmt.Sprintf("%s should be a Base64 String", field)
	case "base64url":
		errMessage = fmt.Sprintf("%s should be a Base64URL String", field)
	case "base64rawurl":
		errMessage = fmt.Sprintf("%s should be a Base64RawURL String", field)
	case "mongodb":
		errMessage = fmt.Sprintf("%s should be a MongoDB ObjectID", field)
	case "datetime":
		errMessage = fmt.Sprintf("%s should be a Datetime", field)
	case "timezone":
		errMessage = fmt.Sprintf("%s should be a Timezone", field)
	case "uuid":
		errMessage = fmt.Sprintf("%s should be a Universally Unique Identifier (UUID)", field)
	case "md4":
		errMessage = fmt.Sprintf("%s should be an MD4 hash", field)
	case "md5":
		errMessage = fmt.Sprintf("%s should be an MD5 hash", field)
	case "sha256":
		errMessage = fmt.Sprintf("%s should be a SHA256 hash", field)
	case "sha384":
		errMessage = fmt.Sprintf("%s should be a SHA384 hash", field)
	default:
		errMessage = fmt.Sprintf("Validation error in field '%s' with tag: %s", field, tag)
	}
	return
}
