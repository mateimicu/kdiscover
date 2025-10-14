// Package cmd offers CLI functionality
package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mateimicu/kdiscover/internal/cluster"
	"github.com/mateimicu/kdiscover/internal/kubeconfig"
)

func Test_generateBackupNameNoConflict(t *testing.T) {
	dir, err := ioutil.TempDir("", ".kube")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.RemoveAll(dir)

	kubeconfigPath := filepath.Join(dir, "kubeconfig")
	backupKubeconfigPath := filepath.Join(dir, "kubeconfig.bak")

	if err := ioutil.WriteFile(kubeconfigPath, []byte("..."), 0600); err != nil {
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
	dir, err := ioutil.TempDir("", "dir")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.RemoveAll(dir)
	if fileExists(dir) {
		t.Errorf("Return true on dir %v", dir)
	}
}

func Test_fileExistsMissing(t *testing.T) {
	dir, err := ioutil.TempDir("", "dir")
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
	dir, err := ioutil.TempDir("", "dir")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "kubeconfig")

	if err := ioutil.WriteFile(path, []byte("...\n"), 0600); err != nil {
		t.Error(err.Error())
	}

	if !fileExists(path) {
		t.Errorf("Return false on file %v", path)
	}
}

func Test_onlyNewFlagLogic(t *testing.T) {
	// Create test clusters
	mockClusters := cluster.GetPredictableMockClusters(3)
	
	// Create a temporary kubeconfig file
	dir, err := ioutil.TempDir("", "kubeconfig-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	
	kubeconfigPath := filepath.Join(dir, "config")
	
	// Create a kubeconfig with one cluster already exported
	kc := kubeconfig.New()
	kc.AddCluster(mockClusters[0], "existing-cluster")
	if err := kc.Persist(kubeconfigPath); err != nil {
		t.Fatal(err)
	}
	
	// Load the kubeconfig back
	loadedKc, err := kubeconfig.LoadKubeconfig(kubeconfigPath)
	if err != nil {
		t.Fatal(err)
	}
	
	// Test that first cluster is exported, others are not
	if !loadedKc.IsExported(mockClusters[0]) {
		t.Error("Expected first cluster to be exported")
	}
	if loadedKc.IsExported(mockClusters[1]) {
		t.Error("Expected second cluster to not be exported")
	}
	if loadedKc.IsExported(mockClusters[2]) {
		t.Error("Expected third cluster to not be exported")
	}
	
	// Test the filtering logic (what would happen with onlyNew flag)
	exportedCount := 0
	newCount := 0
	
	for _, cls := range mockClusters {
		if loadedKc.IsExported(cls) {
			exportedCount++
		} else {
			newCount++
		}
	}
	
	if exportedCount != 1 {
		t.Errorf("Expected 1 exported cluster, got %d", exportedCount)
	}
	if newCount != 2 {
		t.Errorf("Expected 2 new clusters, got %d", newCount)
	}
}
