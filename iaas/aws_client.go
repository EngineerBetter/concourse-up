package iaas

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
)

// AWSClient is the concrete implementation of IClient on AWS
type AWSClient struct {
	region string
}

// NewAWS returns a new AWS client
func NewAWS(region string) (IClient, error) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		return nil, errors.New("env var AWS_ACCESS_KEY_ID not found")
	}

	if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		return nil, errors.New("env var AWS_SECRET_ACCESS_KEY not found")
	}

	return &AWSClient{region}, nil
}

// Region returns the region to operate against
func (client *AWSClient) Region() string {
	return client.region
}

// IAAS returns the iaas to operate against
func (client *AWSClient) IAAS() string {
	return "AWS"
}

// DeleteVMsInVPC deletes all the VMs in the given VPC
func (client *AWSClient) DeleteVMsInVPC(vpcID string) error {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return err
	}

	filterName := "vpc-id"
	ec2Client := ec2.New(sess, &aws.Config{Region: &client.region})

	resp, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: &filterName,
				Values: []*string{
					&vpcID,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	instancesToTerminate := []*string{}
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			fmt.Printf("Terminating instance %s\n", *instance.InstanceId)
			instancesToTerminate = append(instancesToTerminate, instance.InstanceId)
		}
	}

	if len(instancesToTerminate) == 0 {
		return nil
	}

	_, err = ec2Client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: instancesToTerminate,
	})
	return err
}

// FindLongestMatchingHostedZone finds the longest hosted zone that matches the given subdomain
func (client *AWSClient) FindLongestMatchingHostedZone(subdomain string) (string, string, error) {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return "", "", err
	}

	r53Client := route53.New(sess)
	hostedZones := []*route53.HostedZone{}
	err = r53Client.ListHostedZonesPages(&route53.ListHostedZonesInput{}, func(output *route53.ListHostedZonesOutput, _ bool) bool {
		hostedZones = append(hostedZones, output.HostedZones...)
		return true
	})
	if err != nil {
		return "", "", err
	}

	longestMatchingHostedZoneName := ""
	longestMatchingHostedZoneID := ""
	for _, hostedZone := range hostedZones {
		domain := strings.TrimRight(*hostedZone.Name, ".")
		id := *hostedZone.Id
		if strings.HasSuffix(subdomain, domain) {
			if len(domain) > len(longestMatchingHostedZoneName) {
				longestMatchingHostedZoneName = domain
				longestMatchingHostedZoneID = id
			}
		}
	}

	if longestMatchingHostedZoneName == "" {
		return "", "", fmt.Errorf("No matching hosted zone found for domain %s", subdomain)
	}

	longestMatchingHostedZoneID = strings.Replace(longestMatchingHostedZoneID, "/hostedzone/", "", -1)

	return longestMatchingHostedZoneName, longestMatchingHostedZoneID, err
}
