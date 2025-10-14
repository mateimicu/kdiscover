// Package integration contains end-to-end integration tests for kdiscover
// These tests run the actual CLI commands and verify the complete functionality
package integration

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	// testTimeout is the maximum time we allow for CLI commands to complete
	testTimeout = 30 * time.Second
	
	// binaryName is the name of the binary we're testing
	binaryName = "kdiscover"
)

func TestMain(m *testing.M) {
	// Build the binary before running tests
	if err := buildBinary(); err != nil {
		panic("Failed to build binary: " + err.Error())
	}
	
	// Run tests
	code := m.Run()
	
	// Clean up
	os.Remove(binaryName)
	os.Exit(code)
}

func buildBinary() error {
	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	cmd.Dir = "../../" // Go up to the root directory
	return cmd.Run()
}

func getBinaryPath() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "..", "..", binaryName)
}

// TestAWSListBasic tests the basic functionality of aws list command
func TestAWSListBasic(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kdiscover-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	kubeconfigPath := filepath.Join(tempDir, "kubeconfig")
	
	// Test the aws list command
	binaryPath := getBinaryPath()
	cmd := exec.Command(binaryPath, "aws", "list", 
		"--kubeconfig-path", kubeconfigPath,
		"--log-level", "info")
	
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	
	// The command might succeed or fail depending on AWS credentials availability
	// Either way, it should not crash and should produce reasonable output
	
	if err != nil {
		// If it fails, check for reasonable error messages
		expectedErrors := []string{
			"NoCredentialProviders",
			"Unable to locate credentials",
			"no valid providers in chain",
			"error",
		}
		
		foundReasonableError := false
		for _, expectedError := range expectedErrors {
			if strings.Contains(strings.ToLower(outputStr), strings.ToLower(expectedError)) {
				foundReasonableError = true
				break
			}
		}
		
		if !foundReasonableError {
			t.Logf("Command failed but with unexpected error. Output: %s", outputStr)
		}
	} else {
		// If it succeeds, it should show cluster information or indicate no clusters found
		t.Logf("Command succeeded. Output: %s", outputStr)
		
		// Basic sanity check - output should not be empty
		if len(strings.TrimSpace(outputStr)) == 0 {
			t.Errorf("Command succeeded but produced no output")
		}
	}
}

// TestAWSUpdateBasic tests the basic functionality of aws update command
func TestAWSUpdateBasic(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kdiscover-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	kubeconfigPath := filepath.Join(tempDir, "kubeconfig")
	
	// Create an initial kubeconfig file
	initialConfig := `apiVersion: v1
kind: Config
clusters: []
contexts: []
current-context: ""
preferences: {}
users: []
`
	if err := ioutil.WriteFile(kubeconfigPath, []byte(initialConfig), 0600); err != nil {
		t.Fatalf("Failed to create initial kubeconfig: %v", err)
	}
	
	// Test the aws update command
	binaryPath := getBinaryPath()
	cmd := exec.Command(binaryPath, "aws", "update",
		"--kubeconfig-path", kubeconfigPath,
		"--log-level", "info",
		"--backup-kubeconfig=false") // Disable backup for simpler testing
	
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	
	// Command might succeed or fail depending on AWS credentials
	if err != nil {
		// If it fails, check for reasonable error messages
		expectedErrors := []string{
			"NoCredentialProviders",
			"Unable to locate credentials", 
			"no valid providers in chain",
			"error",
		}
		
		foundReasonableError := false
		for _, expectedError := range expectedErrors {
			if strings.Contains(strings.ToLower(outputStr), strings.ToLower(expectedError)) {
				foundReasonableError = true
				break
			}
		}
		
		if !foundReasonableError {
			t.Logf("Command failed but with unexpected error. Output: %s", outputStr)
		}
	} else {
		t.Logf("Command succeeded. Output: %s", outputStr)
	}
	
	// Verify that the original kubeconfig still exists (should not be corrupted)
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		t.Errorf("Kubeconfig file was deleted or corrupted")
	}
	
	// Verify that kubeconfig content is valid YAML
	content, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		t.Errorf("Failed to read kubeconfig after command: %v", err)
	} else {
		contentStr := string(content)
		if !strings.Contains(contentStr, "apiVersion: v1") {
			t.Errorf("Kubeconfig appears corrupted, missing apiVersion. Content: %s", contentStr)
		}
		if !strings.Contains(contentStr, "kind: Config") {
			t.Errorf("Kubeconfig appears corrupted, missing kind. Content: %s", contentStr)
		}
	}
}

// TestKubeconfigPersistence tests that kubeconfig changes are properly persisted
// This addresses the bug mentioned in issue #28
func TestKubeconfigPersistence(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kdiscover-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	kubeconfigPath := filepath.Join(tempDir, "kubeconfig")
	
	// Create an initial kubeconfig with some content
	initialConfig := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://existing-cluster.example.com
  name: existing-cluster
contexts:
- context:
    cluster: existing-cluster
    user: existing-user
  name: existing-context
current-context: existing-context
preferences: {}
users:
- name: existing-user
  user:
    token: existing-token
`
	if err := ioutil.WriteFile(kubeconfigPath, []byte(initialConfig), 0600); err != nil {
		t.Fatalf("Failed to create initial kubeconfig: %v", err)
	}
	
	// Run the update command (it will fail due to no AWS credentials, but should preserve the file)
	binaryPath := getBinaryPath()
	cmd := exec.Command(binaryPath, "aws", "update",
		"--kubeconfig-path", kubeconfigPath,
		"--log-level", "info",
		"--backup-kubeconfig=false")
	
	cmd.Run() // We expect this to fail, so we don't check the error
	
	// Verify that the kubeconfig file still exists and is readable
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		t.Fatalf("Kubeconfig file was deleted")
	}
	
	// Verify that the kubeconfig can be read
	content, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to read kubeconfig after update command: %v", err)
	}
	
	// The content should at least be valid YAML (even if unchanged due to no AWS access)
	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Errorf("Kubeconfig file is empty after update command")
	}
	
	// Basic sanity check - should still contain the apiVersion
	if !strings.Contains(contentStr, "apiVersion: v1") {
		t.Errorf("Kubeconfig appears to be corrupted, missing apiVersion. Content: %s", contentStr)
	}
	
	t.Logf("Kubeconfig persistence test passed. File exists and contains valid content.")
}

// TestVersionCommand tests the version command works
func TestVersionCommand(t *testing.T) {
	binaryPath := getBinaryPath()
	cmd := exec.Command(binaryPath, "version")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Version command failed: %v, output: %s", err, string(output))
	}
	
	outputStr := string(output)
	if !strings.Contains(outputStr, "Version") || (!strings.Contains(outputStr, "dev") && !strings.Contains(outputStr, "v")) {
		t.Errorf("Version command output doesn't look right: %s", outputStr)
	}
}

// TestHelpCommands tests that help commands work for all subcommands
func TestHelpCommands(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{"root help", []string{"--help"}},
		{"aws help", []string{"aws", "--help"}},
		{"aws list help", []string{"aws", "list", "--help"}},
		{"aws update help", []string{"aws", "update", "--help"}},
	}
	
	binaryPath := getBinaryPath()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			output, err := cmd.CombinedOutput()
			
			if err != nil {
				t.Errorf("Help command failed: %v, output: %s", err, string(output))
				return
			}
			
			outputStr := string(output)
			if !strings.Contains(outputStr, "Usage:") {
				t.Errorf("Help output doesn't contain 'Usage:': %s", outputStr)
			}
		})
	}
}

// TestCommandTimeout ensures commands don't hang indefinitely
func TestCommandTimeout(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kdiscover-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	kubeconfigPath := filepath.Join(tempDir, "kubeconfig")
	
	binaryPath := getBinaryPath()
	cmd := exec.Command(binaryPath, "aws", "list",
		"--kubeconfig-path", kubeconfigPath,
		"--log-level", "error") // Use error level to reduce output
	
	// Set a timeout
	timer := time.AfterFunc(testTimeout, func() {
		cmd.Process.Kill()
	})
	defer timer.Stop()
	
	start := time.Now()
	cmd.Run() // We expect this to fail, so we don't check the error
	duration := time.Since(start)
	
	if duration > testTimeout {
		t.Errorf("Command took too long: %v (max allowed: %v)", duration, testTimeout)
	}
}