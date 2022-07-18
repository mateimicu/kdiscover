// Package aws provides function for generatig kubeconfig for EKS clusters
package aws

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver"
)

func TestAWSCLIVersion(t *testing.T) {
	t.Parallel()
	tts := []struct {
		Output, Expected, SupportCommand string
	}{
		{"aws-cli/2.0.10 Python/3.8.2 Darwin/19.3.0 botocore/2.0.0dev14", "2.0.10", "echo"},
		{"aws-cli/2.0.0 Python/3.8.2 Darwin/19.3.0 botocore/2.0.0dev14", "2.0.0", "echo"},
		{"aws-cli/2.0.1 Python/3.8.2 Darwin/19.3.0 botocore/2.0.0dev14", "2.0.1", "echo"},
		{"aws-cli/a.b.c Python/3.8.2 Darwin/19.3.0 botocore/2.0.0dev14", "0.0.0", "echo"},
		{"", "0.0.0", "echo"},
		{"", "0.0.0", "failed-cmd"},
	}

	for _, tt := range tts {
		testname := fmt.Sprintf("%v -> %v (cmd %v)", tt.Output, tt.Expected, tt.SupportCommand)
		t.Run(testname, func(t *testing.T) {
			// overwrite command
			out := getAWSCLIversion([]string{tt.SupportCommand, fmt.Sprintf("'%v'", tt.Output)})

			expVer, err := semver.NewVersion(tt.Expected)
			if err != nil {
				t.Errorf("Can't convert expecte %v in semnver", tt.Expected)
			}

			if expVer.String() != out.String() {
				t.Errorf("Expected %v but got %v from %v", tt.Expected, out, awsCLIVersionCommand)
			}
		})
	}
}
