// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Version(t *testing.T) {
	cases := []struct {
		Cmd []string
	}{
		{Cmd: []string{"version"}},
		{Cmd: []string{"version", "-o", "json"}},
		{Cmd: []string{"version", "-o", "yaml"}},
	}
	for _, tt := range cases {
		testname := fmt.Sprintf("command %v", tt.Cmd)
		t.Run(testname, func(t *testing.T) {
			dir, err := ioutil.TempDir("", ".kube")
			if err != nil {
				t.Error(err.Error())
			}
			defer os.RemoveAll(dir)

			kubeconfigPath := filepath.Join(dir, "kubeconfig")
			version := "mock-version"
			commit := "mock-commit"
			date := "mock-date"
			cmd := NewRootCommand(version, commit, date, "kdisocver")
			buf := new(strings.Builder)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			completCmd := append(tt.Cmd, "--kubeconfig-path")
			completCmd = append(completCmd, kubeconfigPath)

			cmd.SetArgs(completCmd)
			err = cmd.Execute()
			if err != nil {
				t.Error(err.Error())
			}
			out := buf.String()
			assert.Contains(t, out, version)
			assert.Contains(t, out, date)
			assert.Contains(t, out, date)
		})
	}
}
