// Package internal provides function for working with EKS cluseters
package cluster

import (
	"fmt"
	"math/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func randomString(length int) string {
	return StringWithCharset(length, charset)
}

func GetMockClusters(c int) []Cluster {
	d := make([]Cluster, 0, c)
	for i := 0; i < c; i++ {
		r := randomString(10)
		d = append(d, Cluster{
			Name:                     fmt.Sprintf("clucster-name-%v-%v", r, i),
			Region:                   fmt.Sprintf("clucster-region-%v-%v", r, i),
			Id:                       fmt.Sprintf("clucster-id-%v-%v", r, i),
			Status:                   fmt.Sprintf("clucster-status-%v-%v", r, i),
			Endpoint:                 fmt.Sprintf("clucster-endpoint-%v-%v", r, i),
			CertificateAuthorityData: fmt.Sprintf("clucster-certificate-authority-data-%v-%v", r, i),
		})
	}
	return d
}
