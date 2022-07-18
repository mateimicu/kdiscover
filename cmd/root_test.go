// Package cmd offers CLI functionality
package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

var update = flag.Bool("update", false, "update .golden files")

var basicCommands []struct{ cmd []string } = []struct {
	cmd []string
}{
	{[]string{"version"}},
	{[]string{"aws"}},
	{[]string{"aws", "list"}},
	{[]string{"aws", "update"}},
}

// Because cobra is not running PersistantPreRunE for all the commands
// until the leaf command we implemented a hack. This test needs to be update
// with all the possible combination of commands in order to check that the logging hack work
// An issue about this https://github.com/spf13/cobra/issues/252
func Test_CascadingPersistPreRunEHackWithLoggingLevels(t *testing.T) {
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
				cmd := NewRootCommand("", "", "", "kdiscover")
				cmd.SetOut(ioutil.Discard)
				cmd.SetErr(ioutil.Discard)

				completCmd := append(tt.cmd, "--log-level")
				completCmd = append(completCmd, k)
				completCmd = append(completCmd, "--kubeconfig-path")
				completCmd = append(completCmd, kubeconfigPath)

				cmd.SetArgs(completCmd)
				err = cmd.Execute()
				if err != nil {
					t.Error(err.Error())
				}

				// none logging level is a special case
				if k == "none" {
					if log.StandardLogger().Out != ioutil.Discard {
						t.Errorf("Running %v we were expecting logging to be discared but it is not ", completCmd)
					}
				} else {
					if exp != log.GetLevel() {
						t.Errorf("Running %v we were expecting logger to be %v but it is %v", completCmd, exp, log.GetLevel())
					}
				}
			})
		}
	}
}

// This is a smoke test to make sure all commands are able to function
func Test_HelpFunction(t *testing.T) {
	expected := "kdiscover"
	for _, tt := range basicCommands {
		testname := fmt.Sprintf("command %v", tt.cmd)
		t.Run(testname, func(t *testing.T) {
			cmd := NewRootCommand("", "", "", "kdiscover")

			buf := new(strings.Builder)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			completCmd := append(tt.cmd, "--help")

			cmd.SetArgs(completCmd)
			err := cmd.Execute()
			if err != nil {
				t.Error(err.Error())
			}

			if !strings.Contains(buf.String(), expected) {
				t.Errorf(
					"Running %v we were expecting %v in the output but got: %v",
					completCmd, expected, buf.String())
			}
		})
	}
}

func Test_getAllLogglingLevels(t *testing.T) {
	for _, lvl := range getAllLogglingLevels() {
		if _, ok := loggingLevels[lvl]; !ok {
			t.Errorf("Logging level %v not found in map", lvl)
		}
	}
}
