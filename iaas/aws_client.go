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

// IEC2 only implements functions used in the iaas package
type IEC2 interface {
	DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error)
	DeleteVolume(input *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error)
}

func newAWS(region string) (IClient, error) {
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

// NewEC2Client creates a new EC2 client
func (client *AWSClient) NewEC2Client() (IEC2, error) {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return nil, err
	}
	ec2Client := ec2.New(sess, &aws.Config{Region: &client.region})
	return ec2Client, nil
}

// DeleteVolumes deletes the specified EBS volumes
func (client *AWSClient) DeleteVolumes(volumes []*string, deleteVolume func(ec2Client IEC2, volumeID *string) error, newEC2Client func() (IEC2, error)) error {
	if len(volumes) == 0 {
		return nil
	}

	ec2Client, err := newEC2Client()
	if err != nil {
		return err
	}

	volumesOutput, err := ec2Client.DescribeVolumes(&ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("status"),
				Values: []*string{
					aws.String("available"),
				},
			},
		},
		VolumeIds: volumes,
	})
	if err != nil {
		return err
	}

	volumesToDelete := volumesOutput.Volumes

	for _, volume := range volumesToDelete {
		volumeID := volume.VolumeId
		err = deleteVolume(ec2Client, volumeID)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteVolume deletes an EBS volume with the given ID
func DeleteVolume(ec2Client IEC2, volumeID *string) error {
	fmt.Printf("Deleting volume: %s\n", *volumeID)
	_, err := ec2Client.DeleteVolume(&ec2.DeleteVolumeInput{
		VolumeId: volumeID,
	})
	return err
}

// DeleteVMsInVPC deletes all the VMs in the given VPC
func (client *AWSClient) DeleteVMsInVPC(vpcID string) ([]*string, error) {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return nil, err
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
		return nil, err
	}

	instancesToTerminate := []*string{}
	volumesToDelete := []*string{}
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			fmt.Printf("Terminating instance %s\n", *instance.InstanceId)
			instancesToTerminate = append(instancesToTerminate, instance.InstanceId)
			for _, blockDevice := range instance.BlockDeviceMappings {
				volumesToDelete = append(volumesToDelete, blockDevice.Ebs.VolumeId)
			}
		}
	}

	if len(instancesToTerminate) == 0 {
		return nil, nil
	}

	_, err = ec2Client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: instancesToTerminate,
	})
	if err != nil {
		return nil, err
	}

	return volumesToDelete, nil
}

// ListHostedZones returns a list of hosted zones
func ListHostedZones() ([]*route53.HostedZone, error) {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return nil, err
	}

	r53Client := route53.New(sess)
	hostedZones := []*route53.HostedZone{}
	err = r53Client.ListHostedZonesPages(&route53.ListHostedZonesInput{}, func(output *route53.ListHostedZonesOutput, _ bool) bool {
		hostedZones = append(hostedZones, output.HostedZones...)
		return true
	})
	if err != nil {
		return nil, err
	}

	return hostedZones, nil
}

// FindLongestMatchingHostedZone finds the longest hosted zone that matches the given subdomain
func (client *AWSClient) FindLongestMatchingHostedZone(subdomain string, listHostedZones func() ([]*route53.HostedZone, error)) (string, string, error) {
	hostedZones, err := listHostedZones()
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
