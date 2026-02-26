// Package cmd offers CLI functionality
package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_generateBackupNameNoConflict(t *testing.T) {
	dir, err := os.MkdirTemp("", ".kube")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.RemoveAll(dir)

	kubeconfigPath := filepath.Join(dir, "kubeconfig")
	backupKubeconfigPath := filepath.Join(dir, "kubeconfig.bak")

	if err := os.WriteFile(kubeconfigPath, []byte("..."), 0600); err != nil {
		t.Error(err.Error())
	}

	bName, err := backupKubeConfig(kubeconfigPath)
	if err != nil {
		t.Error(err.Error())
	}

	if !fileExists(backupKubeconfigPath) {
		t.Errorf("Expecing %v to exist as backup of %v", backupKubeconfigPath, kubeconfigPath)
	}

	if bName != backupKubeconfigPath {
		t.Errorf("Backup name is %v, expected %v", bName, backupKubeconfigPath)
	}
}

func Test_fileExistsDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "dir")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.RemoveAll(dir)
	if fileExists(dir) {
		t.Errorf("Return true on dir %v", dir)
	}
}

func Test_fileExistsMissing(t *testing.T) {
	dir, err := os.MkdirTemp("", "dir")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "missing")
	if fileExists(dir) {
		t.Errorf("Return true on missing file %v", path)
	}
}

func Test_fileExistsFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "dir")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "kubeconfig")

	if err := os.WriteFile(path, []byte("...\n"), 0600); err != nil {
		t.Error(err.Error())
	}

	if !fileExists(path) {
		t.Errorf("Return false on file %v", path)
	}
}
