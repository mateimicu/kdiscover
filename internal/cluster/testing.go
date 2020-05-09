// Package internal provides function for working with EKS cluseters
package cluster

import (
	"fmt"
	"math/rand"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	lenID   = 10
)

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func randomString(length int) string {
	return stringWithCharset(length, charset)
}

func GetMockClusters(c int) []*Cluster {
	d := make([]*Cluster, 0, c)
	for i := 0; i < c; i++ {
		r := randomString(lenID)

		c := NewCluster()
		c.Name = fmt.Sprintf("clucster-name-%v-%v", r, i)
		c.Region = fmt.Sprintf("clucster-region-%v-%v", r, i)
		c.ID = fmt.Sprintf("clucster-id-%v-%v", r, i)
		c.Status = fmt.Sprintf("clucster-status-%v-%v", r, i)
		c.Endpoint = fmt.Sprintf("clucster-endpoint-%v-%v", r, i)
		c.CertificateAuthorityData = fmt.Sprintf("clucster-certificate-authority-data-%v-%v", r, i)

		d = append(d, c)
	}
	return d
}
