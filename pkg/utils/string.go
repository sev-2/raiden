package utils

import (
	"regexp"
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
	reg := regexp.MustCompile("[^a-zA-Z]+")
	result := reg.ReplaceAllString(s, "")
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
