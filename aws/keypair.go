package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// AssertAWSKeyPairExists  checks if an AWS keypair exists and creates one if not (or if the given private key is empty)
func AssertAWSKeyPairExists(awsClient *ec2.EC2, keyPairName string, privateKeyContents string) (string, error) {
	filterName := "key-name"
	describeKeyPairsOutput, err := awsClient.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   &filterName,
				Values: []*string{&keyPairName},
			},
		},
	})
	if err != nil {
		return "", err
	}

	keyPairExists := len(describeKeyPairsOutput.KeyPairs) != 0

	// generate a new key if one doesn't exist on AWS or if we don't have the private key in the credentials file
	if keyPairExists && privateKeyContents != "" {
		fmt.Printf("Key pair %s already exists in AWS and in local config\n", keyPairName)
		return privateKeyContents, nil
	}

	if keyPairExists {
		fmt.Printf("Key pair %s found in aws but not locally\nDeleting existing key pair\n", keyPairName)
		if _, err := awsClient.DeleteKeyPair(&ec2.DeleteKeyPairInput{KeyName: &keyPairName}); err != nil {
			return "", err
		}
	}

	fmt.Printf("Creating new key pair %s\n", keyPairName)
	createKeyPairOutput, err := awsClient.CreateKeyPair(&ec2.CreateKeyPairInput{KeyName: &keyPairName})
	if err != nil {
		return "", err
	}

	return *createKeyPairOutput.KeyMaterial, nil
}
