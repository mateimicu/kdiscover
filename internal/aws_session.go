// Package internal provides wrapper for creating aws sessions
package internal

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
)

func getAWSSession(region string) (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	return sess, err
}

func getEKSClient(session *session.Session) *eks.EKS {
	return eks.New(session)
}
