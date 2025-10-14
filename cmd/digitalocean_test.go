// Package cmd offers CLI functionality
package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestDigitalOceanCommand(t *testing.T) {
	cmd := NewRootCommand("", "", "", "kdiscover")
	buf := new(strings.Builder)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	log.SetOutput(io.Discard)
	defer func() { log.SetOutput(os.Stdout) }()
	hook := test.NewGlobal()
	defer hook.Reset()

	// Test help command
	cmd.SetArgs([]string{"digitalocean", "--help"})
	err := cmd.Execute()
	if err != nil {
		t.Error(err.Error())
	}

	// Verify output contains expected elements
	output := buf.String()
	if !strings.Contains(output, "Work with DigitalOcean DOKS clusters") {
		t.Error("Expected help output to contain DigitalOcean description")
	}
	if !strings.Contains(output, "digitalocean, do") {
		t.Error("Expected help output to contain aliases")
	}
	if !strings.Contains(output, "list") {
		t.Error("Expected help output to contain list command")
	}
	if !strings.Contains(output, "update") {
		t.Error("Expected help output to contain update command")
	}
}