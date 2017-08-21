package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func DeleteVMsInVPC(vpcID string, region string) error {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return err
	}

	filterName := "vpc-id"
	client := ec2.New(sess, &aws.Config{Region: &region})

	resp, err := client.DescribeInstances(&ec2.DescribeInstancesInput{
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

	_, err = client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: instancesToTerminate,
	})
	return err
}
