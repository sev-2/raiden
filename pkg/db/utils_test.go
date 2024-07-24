package db

import (
	"strings"
	"testing"
)

func TestKeyExist(t *testing.T) {
	input := make(map[string]string)
	input["foo"] = "foo"
	input["bar"] = "bar"
	input["baz"] = "baz"

	if !keyExist(input, "foo") {
		t.Errorf("Expected \"%s\" is exist", "foo")
	}

	if keyExist(input, "qux") {
		t.Errorf("Expected \"%s\" is not exist", "qux")
	}
}

func TestReverseSortString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "User.Team.Orgs",
			expected: "Orgs.Team.User",
		},
		{
			input:    "User.Comment.ArticleMockModel",
			expected: "ArticleMockModel.Comment.User",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			input := strings.Split(tc.input, ".")
			actual := strings.Join(reverseSortString(input), ".")

			if actual != tc.expected {
				t.Errorf("Expected: %s, Actual: %s", tc.expected, actual)
			}
		})
	}
}
