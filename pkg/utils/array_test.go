package utils_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRemoveByIndex(t *testing.T) {
	tests := []struct {
		name     string
		source   []int
		index    []int
		expected []int
	}{
		{
			name:     "Remove single element",
			source:   []int{1, 2, 3, 4, 5},
			index:    []int{2},
			expected: []int{1, 2, 4, 5},
		},
		{
			name:     "Remove multiple elements",
			source:   []int{1, 2, 3, 4, 5},
			index:    []int{1, 3},
			expected: []int{1, 3, 5},
		},
		{
			name:     "Remove first and last elements",
			source:   []int{1, 2, 3, 4, 5},
			index:    []int{0, 4},
			expected: []int{2, 3, 4},
		},
		{
			name:     "Remove no elements",
			source:   []int{1, 2, 3, 4, 5},
			index:    []int{},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Remove all elements",
			source:   []int{1, 2, 3, 4, 5},
			index:    []int{0, 1, 2, 3, 4},
			expected: []int(nil),
		},
		{
			name:     "Remove with out-of-bound index",
			source:   []int{1, 2, 3, 4, 5},
			index:    []int{5, 6},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Empty source slice",
			source:   []int{},
			index:    []int{0, 1},
			expected: []int(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.RemoveByIndex(tt.source, tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}
