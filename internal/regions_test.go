package internal

import (
	"fmt"
	"testing"
)

func TestContains(t *testing.T) {
	tts := []struct {
		key      string
		list     []string
		expected bool
	}{
		{"to_be_found", []string{"to_be_found", "another_item"}, true},
		{"not_in_list", []string{"item", "another_item"}, false},
		{"not_in_list", []string{}, false},
	}

	for _, tt := range tts {
		testname := fmt.Sprintf("%v in %v", tt.key, tt.list)
		t.Run(testname, func(t *testing.T) {
			result := contains(tt.key, tt.list)
			if result != tt.expected {
				t.Errorf("contains of %v in %v is incorrect, got: %v, want: %v.", tt.key, tt.list, result, tt.expected)
			}
		})
	}
}
