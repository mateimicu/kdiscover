// Package cmd offers CLI functionality
package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_generateBackupNameNoConflict(t *testing.T) {
	dir, err := ioutil.TempDir("", ".kube")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.RemoveAll(dir)

	kubeconfigPath := filepath.Join(dir, "kubeconfig")
	backupKubeconfigPath := filepath.Join(dir, "kubeconfig.bak")

	if err := ioutil.WriteFile(kubeconfigPath, []byte("..."), 0666); err != nil {
		t.Error(err.Error())
	}

	err = backupKubeConfig(kubeconfigPath)
	if err != nil {
		t.Error(err.Error())
	}

	if !fileExists(backupKubeconfigPath) {
		t.Errorf("Expecing %v to exist as backup of %v", backupKubeconfigPath, kubeconfigPath)
	}
}
