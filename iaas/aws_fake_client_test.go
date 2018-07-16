package iaas_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type fakeEC2Client struct {
}

func (fakeEC2Client *fakeEC2Client) DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	var inputVolumes []*string
	for _, filter := range input.Filters {
		if *filter.Name == "volume-id" {
			inputVolumes = filter.Values
			break
		}
	}
	volume1 := &ec2.Volume{
		VolumeId: inputVolumes[0],
	}
	volume2 := &ec2.Volume{
		VolumeId: inputVolumes[1],
	}
	volumes := []*ec2.Volume{volume1, volume2}
	output := &ec2.DescribeVolumesOutput{
		Volumes: volumes,
	}
	return output, nil
}

func (fakeEC2Client *fakeEC2Client) DeleteVolume(input *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error) {
	return nil, nil
}

func (fakeEC2Client *fakeEC2Client) DescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	ipRange1 := &ec2.IpRange{
		CidrIp:      aws.String("1.2.3.4"),
		Description: aws.String("present"),
	}
	ipRange2 := &ec2.IpRange{
		CidrIp:      aws.String("5.6.7.8"),
		Description: aws.String("invalid"),
	}

	rangesSingle := []*ec2.IpRange{ipRange1}
	rangesDouble := []*ec2.IpRange{ipRange1, ipRange2}

	fromPort22 := int64(22)
	fromPort6868 := int64(6868)
	fromPort25555 := int64(25555)

	entry22 := &ec2.IpPermission{
		FromPort: &fromPort22,
		IpRanges: rangesSingle,
	}
	entry6868 := &ec2.IpPermission{
		FromPort: &fromPort6868,
		IpRanges: rangesDouble,
	}
	entry25555 := &ec2.IpPermission{
		FromPort: &fromPort25555,
		IpRanges: rangesDouble,
	}

	ingressPermissions := []*ec2.IpPermission{
		entry22, entry6868, entry25555,
	}

	securityGroup := &ec2.SecurityGroup{
		IpPermissions: ingressPermissions,
	}

	securityGroups := []*ec2.SecurityGroup{
		securityGroup,
	}

	output := &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: securityGroups,
	}
	return output, nil
}
