// Package cmd offers CLI functionality
package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mateimicu/kdiscover/internal"
)

//import (
//"fmt"
//"io/ioutil"
//"strings"
//"testing"

//log "github.com/sirupsen/logrus"
//"github.com/zenizh/go-capturer"
//)

func getMockClusters(c int) []internal.Cluster {
	d := make([]internal.Cluster, 0, c)
	for i := 0; i < c; i++ {
		d = append(d, internal.Cluster{
			Name:   fmt.Sprintf("clucster-name-%v", i),
			Region: fmt.Sprintf("clucster-region-%v", i),
			Id:     fmt.Sprintf("clucster-id-%v", i),
			Status: fmt.Sprintf("clucster-status-%v", i),
		})
	}
	return d
}

// Test if the number of clusters is corectly diplayed
func Test_getTable(t *testing.T) {
	tts := []struct {
		clusters []internal.Cluster
	}{
		{clusters: getMockClusters(0)},
		{clusters: getMockClusters(1)},
		{clusters: getMockClusters(3)},
	}
	for _, tt := range tts {
		testname := fmt.Sprintf("Clusters %v", tt.clusters)
		t.Run(testname, func(t *testing.T) {
			r := getTable(tt.clusters)
			if !strings.Contains(r, fmt.Sprintf("%v", len(tt.clusters))) {
				t.Errorf("Expected %v in output, but got %v", len(tt.clusters), r)
			}
		})
	}
}
