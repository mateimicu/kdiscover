// Package internal provides function for working with EKS cluseters
package cluster

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func tmpFile(t *testing.T) (string, func()) {
	file, err := ioutil.TempFile("", "prefix")
	if err != nil {
		t.Error(err)
	}

	return file.Name(), func() { os.Remove(file.Name()) }
}

type mockContextNameGenerator struct{}

func (c mockContextNameGenerator) GetContextName(cls Cluster) (string, error) {
	return fmt.Sprintf("ctx-name-%v-%v", cls.Name, cls.Id), nil
}

//func TestIsExported(t *testing.T) {
//t.Parallel()
//cases := map[string]struct {
//// cluster to check if it is exported
//cls Cluster

//// how many existing clusters are in the file (except cls)
//existingClusterCount int

//// exclude Cluster information for cls from the file
//excludeCluster bool

//// exclude Context information for cls from the file
//excludeContext bool

//// the expected output of IsExported
//expected bool
//}{
//"is_the_file": {GetMockClusters(1)[0], 0, false, false, false},
//}

//for testName, tc := range cases {
//t.Run(testName, func(t *testing.T) {
//allClusters := GetMockClusters(0)
//f, cleanup := tmpFile(t)
//defer cleanup()
//if tc.expected {
//allClusters = append(allClusters, tc.cls)
//}
//err := UpdateKubeconfig(allClusters, f, mockContextNameGenerator{})
//if err != nil {
//t.Error(err)
//}

//if tc.cls.IsExported(f) != tc.expected {
//t.Errorf("Expecting IsExported=%v but got %v", tc.cls.IsExported(f), tc.expected)
//}
//})
//}
//}
