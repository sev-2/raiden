package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/gertd/go-pluralize"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func IsStringContainSpace(input string) bool {
	for _, char := range input {
		if unicode.IsSpace(char) {
			return true
		}
	}
	return false
}

func ToSnakeCase(s string) string {
	var result strings.Builder
	result.Grow(len(s) + 5) // Preallocate space for a few extra characters

	for i, char := range s {
		switch {
		case char >= 'A' && char <= 'Z':
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(char + ('a' - 'A'))
		case char == '-':
			result.WriteRune('_')
		default:
			result.WriteRune(char)
		}
	}

	return result.String()
}

func ToGoModuleName(s string) string {
	var moduleName string

	splitedPath := strings.Split(strings.TrimRight(s, "/"), "/")
	if len(splitedPath) == 1 {
		moduleName = splitedPath[0]
	} else if len(splitedPath) > 0 {
		moduleName = splitedPath[len(splitedPath)-1]
	}

	reg := regexp.MustCompile("[^a-zA-Z]+")
	result := reg.ReplaceAllString(moduleName, "")

	return strings.ToLower(result)
}

func SnakeCaseToPascalCase(s string) string {
	words := strings.Split(s, "_")
	caser := cases.Title(language.Und, cases.NoLower)

	for i := range words {
		words[i] = caser.String(words[i])
	}
	return strings.Join(words, "")
}

func ToPlural(word string) string {
	client := pluralize.NewClient()
	return client.Plural(word)
}

func MatchReplacer(query string, paramKey string, replacement string) string {
	// Split the query into individual words
	words := strings.Fields(query)

	// Replace the parameter in the query, considering context
	for i := 0; i < len(words); i++ {
		if words[i] == paramKey {
			// Check if there is a dot (.) before or after the parameter
			hasDotBefore := i > 0 && strings.HasSuffix(words[i-1], ".")
			hasDotAfter := i < len(words)-1 && strings.HasPrefix(words[i+1], ".")

			// Replace only if it's not part of a longer identifier (e.g., :u.Role)
			if !hasDotBefore && !hasDotAfter {
				words[i] = replacement
			}
		}
	}

	// Join the words back into a string
	return strings.Join(words, " ")
}

func CleanUpString(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			s, "\t", " ",
		),
		"\n", " ",
	)
}

func HashString(query string) string {
	return HashByte([]byte(query))
}

func HashByte(queryByte []byte) string {
	hasher := sha256.New()
	hasher.Write(queryByte)
	return hex.EncodeToString(hasher.Sum(nil))
}

// RemoveParenthesesContent removes content within parentheses from a string.
func RemoveParenthesesContent(input string) string {
	re := regexp.MustCompile(`\([^)]*\)`)
	return re.ReplaceAllString(input, "")
}

func ToCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// Convert first character to lowercase
	camel := strings.ToLower(string(s[0]))

	// Convert subsequent uppercase characters to lowercase preceded by an underscore
	for _, r := range s[1:] {
		if unicode.IsUpper(r) {
			camel += "_" + strings.ToLower(string(r))
		} else {
			camel += string(r)
		}
	}

	return camel
}

func ParseTag(rawTag string) map[string]string {
	tagMap := make(map[string]string)
	for _, t := range strings.Split(rawTag, " ") {
		splitTag := strings.SplitN(t, ":", 2)
		if len(splitTag) == 2 {
			tagMap[splitTag[0]] = strings.Trim(splitTag[1], "\"")
		}
	}
	return tagMap
}

func ParseBool(str string) bool {
	val, _ := strconv.ParseBool(str)
	return val
}
