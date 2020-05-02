// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zenizh/go-capturer"
)

var basicCommands []struct{ cmd []string } = []struct {
	cmd []string
}{
	{[]string{"aws"}},
	{[]string{"aws", "list"}},
	{[]string{"aws", "update"}},
}

// Because cobra is not runing PersistantPreRunE for all the commands
// untill the leaf command we implemented a hack. This test needs to be update
// with all the possible combination of commands in order to check that the logging hack worsk
// An issue about this https://github.com/spf13/cobra/issues/252
func Test_CascadingPersistPreRunEHackWithLoggingLevels(t *testing.T) {
	for _, tt := range basicCommands {
		for k, exp := range loggingLevels {

			testname := fmt.Sprintf("command %v and logging lvl %v", tt.cmd, k)
			t.Run(testname, func(t *testing.T) {
				cmd := NewRootCommand()
				cmd.SetOut(ioutil.Discard)
				cmd.SetErr(ioutil.Discard)

				completCmd := append(tt.cmd, "--log-level")
				completCmd = append(completCmd, k)

				cmd.SetArgs(completCmd)
				capturer.CaptureOutput(func() {
					cmd.Execute()
				})

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
			cmd := NewRootCommand()

			completCmd := append(tt.cmd, "--help")

			cmd.SetArgs(completCmd)
			out := capturer.CaptureOutput(func() {
				cmd.Execute()
			})

			if !strings.Contains(string(out), expected) {
				t.Errorf("Running %v we were expecting %v in the ouput but got: %v", completCmd, expected, out)
			}
		})
	}
}
