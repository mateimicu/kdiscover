// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	"io/ioutil"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zenizh/go-capturer"
)

// Because cobra is not runing PersistantPreRunE for all the commands
// untill the leaf command we implemented a hack. This test needs to be update
// with all the possible combination of commands in order to check that the logging hack worsk
// An issue about this https://github.com/spf13/cobra/issues/252
func Test_CascadingPersistPreRunEHackWithLoggingLevels(t *testing.T) {
	tts := []struct {
		cmd []string
	}{
		{[]string{"aws"}},
		{[]string{"aws", "list"}},
		{[]string{"aws", "update"}},
	}

	for _, tt := range tts {
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
