// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mateimicu/kdiscover/internal/aws"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

type testCase struct {
	Cmd        []string
	Partitions []string
}

var cases []testCase = []testCase{
	{[]string{"aws", "list"}, []string{"aws"}},
	{[]string{"aws", "list"}, []string{"aws-cn"}},
	{[]string{"aws", "list"}, []string{"aws-iso-b"}},
	{[]string{"aws", "list"}, []string{"aws", "aws-cn"}},
	{[]string{"aws", "list"}, []string{"aws-us-gov", "aws-iso", "aws-iso-b"}},
	{[]string{"aws", "list"}, []string{"aws", "aws-cn", "aws-us-gov", "aws-iso", "aws-iso-b"}},
	{[]string{"aws", "update"}, []string{"aws"}},
	{[]string{"aws", "update"}, []string{"aws-cn"}},
	{[]string{"aws", "update"}, []string{"aws-iso-b"}},
	{[]string{"aws", "update"}, []string{"aws", "aws-cn"}},
	{[]string{"aws", "update"}, []string{"aws-us-gov", "aws-iso", "aws-iso-b"}},
	{[]string{"aws", "update"}, []string{"aws", "aws-cn", "aws-us-gov", "aws-iso", "aws-iso-b"}},
}

func TestQueryAllRegions(t *testing.T) {
	for _, tt := range cases {
		testname := fmt.Sprintf("command %v", tt.Partitions)
		t.Run(testname, func(t *testing.T) {
			dir, err := ioutil.TempDir("", ".kube")
			if err != nil {
				t.Error(err.Error())
			}
			defer os.RemoveAll(dir)
			kubeconfigPath := filepath.Join(dir, "kubeconfig")

			cmd := NewRootCommand()
			buf := new(strings.Builder)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			log.SetOutput(ioutil.Discard)
			defer func() { log.SetOutput(os.Stdout) }()
			hook := test.NewGlobal()
			defer hook.Reset()
			args := append(tt.Cmd, []string{
				"--log-level", "debug",
				"--kubeconfig-path", kubeconfigPath,
				"--aws-partitions", strings.Join(tt.Partitions, ","),
			}...)

			cmd.SetArgs(args)
			err = cmd.Execute()
			if err != nil {
				t.Error(err.Error())
			}

			expectedLogs := make(map[string]bool)
			for _, region := range aws.GetRegions(tt.Partitions) {
				expectedLogs[region] = false
			}
			for _, e := range hook.AllEntries() {
				if v, ok := e.Data["region"]; ok {
					expectedLogs[fmt.Sprintf("%v", v)] = true
				}
			}

			for k, v := range expectedLogs {
				if !v {
					t.Errorf("Could not find log for %v", k)
				}
			}
		})
	}
}
