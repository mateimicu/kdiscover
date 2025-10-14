// Package api contains API tests for the kdiscover CLI commands
// These tests exercise the public CLI interface, as opposed to internal implementation details
package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mateimicu/kdiscover/cmd"
	"github.com/mateimicu/kdiscover/internal/aws"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

type testCase struct {
	Cmd        []string
	Partitions []string
}

var cases = []testCase{
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

// TestQueryAllRegions tests that all expected regions are queried when running CLI commands
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

			command := cmd.NewRootCommand("", "", "", "kdiscover")
			buf := new(strings.Builder)
			command.SetOut(buf)
			command.SetErr(buf)
			log.SetOutput(ioutil.Discard)
			defer func() { log.SetOutput(os.Stdout) }()
			hook := test.NewGlobal()
			defer hook.Reset()
			tt.Cmd = append(tt.Cmd, []string{
				"--log-level", "debug",
				"--kubeconfig-path", kubeconfigPath,
				"--aws-partitions", strings.Join(tt.Partitions, ","),
			}...)

			command.SetArgs(tt.Cmd)
			err = command.Execute()
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

var basicCommands = []struct {
	cmd     []string
	context string
}{
	{[]string{"version"}, "kdiscover"},
	{[]string{"aws"}, "kdiscover"},
	{[]string{"aws", "list"}, "kdiscover"},
	{[]string{"aws", "update"}, "kdiscover"},
	{[]string{"version"}, "kubectl-discover"},
	{[]string{"aws"}, "kubectl-discover"},
	{[]string{"aws", "list"}, "kubectl-discover"},
	{[]string{"aws", "update"}, "kubectl-discover"},
}

var loggingLevels = map[string]log.Level{
	"panic": log.PanicLevel,
	"fatal": log.FatalLevel,
	"error": log.ErrorLevel,
	"warn":  log.WarnLevel,
	"info":  log.InfoLevel,
	"debug": log.DebugLevel,
	"trace": log.TraceLevel,
}

// getAllLogglingLevels returns all supported logging levels
func getAllLogglingLevels() []string {
	// Get all supported logging levels from the map keys plus "none"
	levels := make([]string, 0, len(loggingLevels)+1)
	for k := range loggingLevels {
		levels = append(levels, k)
	}
	levels = append(levels, "none")
	return levels
}

// TestCascadingPersistPreRunEHackWithLoggingLevels tests the logging level configuration
// Because cobra is not running PersistantPreRunE for all the commands
// until the leaf command we implemented a hack. This test needs to be update
// with all the possible combination of commands in order to check that the logging hack work
// An issue about this https://github.com/spf13/cobra/issues/252
func TestCascadingPersistPreRunEHackWithLoggingLevels(t *testing.T) {
	for _, tt := range basicCommands {
		for k, exp := range loggingLevels {
			testname := fmt.Sprintf("command %v and logging lvl %v", tt.cmd, k)
			t.Run(testname, func(t *testing.T) {
				dir, err := ioutil.TempDir("", ".kube")
				if err != nil {
					t.Error(err.Error())
				}
				defer os.RemoveAll(dir)

				kubeconfigPath := filepath.Join(dir, "kubeconfig")
				command := cmd.NewRootCommand("", "", "", tt.context)
				command.SetOut(ioutil.Discard)
				command.SetErr(ioutil.Discard)

				tt.cmd = append(tt.cmd, "--log-level", k, "--kubeconfig-path", kubeconfigPath)

				command.SetArgs(tt.cmd)
				err = command.Execute()
				if err != nil {
					t.Error(err.Error())
				}

				// none logging level is a special case
				if k == "none" {
					if log.StandardLogger().Out != ioutil.Discard {
						t.Errorf("Running %v we were expecting logging to be discared but it is not ", tt.cmd)
					}
				} else {
					if exp != log.GetLevel() {
						t.Errorf("Running %v we were expecting logger to be %v but it is %v", tt.cmd, exp, log.GetLevel())
					}
				}
			})
		}
	}
}

// TestHelpFunction is a smoke test to make sure all commands are able to function
func TestHelpFunction(t *testing.T) {
	for _, tt := range basicCommands {
		testname := fmt.Sprintf("help for command %v", tt.cmd)
		t.Run(testname, func(t *testing.T) {
			dir, err := ioutil.TempDir("", ".kube")
			if err != nil {
				t.Error(err.Error())
			}
			defer os.RemoveAll(dir)

			kubeconfigPath := filepath.Join(dir, "kubeconfig")
			command := cmd.NewRootCommand("", "", "", tt.context)
			command.SetOut(ioutil.Discard)
			command.SetErr(ioutil.Discard)

			// Add help flag and required flags
			helpCmd := append(tt.cmd, "--help", "--kubeconfig-path", kubeconfigPath)

			command.SetArgs(helpCmd)
			err = command.Execute()
			if err != nil {
				t.Error(err.Error())
			}
		})
	}
}

func TestGetAllLogglingLevels(t *testing.T) {
	for _, lvl := range getAllLogglingLevels() {
		if lvl == "none" {
			continue // "none" is a special case not in the map
		}
		if _, ok := loggingLevels[lvl]; !ok {
			t.Errorf("Logging level %v not found in map", lvl)
		}
	}
}