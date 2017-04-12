package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// CreateSession creates a new AWS session
func CreateSession() (*session.Session, error) {
	s, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return nil, err
	}

	return s, nil
}
