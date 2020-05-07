package aws

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	t.Parallel()
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
			assert.Equal(t, result, tt.expected)
		})
	}
}

func TestGetRegions(t *testing.T) {
	t.Parallel()
	tts := []struct {
		partitions []string
	}{
		{[]string{}},
		{[]string{"aws", "aws-cn", "aws-us-gov", "aws-iso", "aws-iso-b"}},
		{[]string{"aws", "aws-cn", "aws-iso-b"}},
		{[]string{"aws-iso", "aws-iso-b"}},
	}

	for _, tt := range tts {
		testname := fmt.Sprintf("Partitions %v", tt.partitions)
		t.Run(testname, func(t *testing.T) {
			totalResult := GetRegions(tt.partitions)

			// compute partial result
			partialResult := make([]string, 0)
			for _, partition := range tt.partitions {
				partialResult = append(partialResult, GetRegions([]string{partition})...)
			}

			sort.Strings(totalResult)
			sort.Strings(partialResult)
			if len(partialResult) == 0 && len(totalResult) == 0 {
				return
			}
			assert.Equal(t, partialResult, totalResult)
		})
	}
}
