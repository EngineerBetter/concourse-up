package iaas_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/EngineerBetter/concourse-up/iaas"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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

var _ = Describe("Client#FindLongestMatchingHostedZone", func() {

	BeforeEach(func() {
		os.Setenv("AWS_SECRET_ACCESS_KEY", "123")
		os.Setenv("AWS_ACCESS_KEY_ID", "123")
	})

	var fakeDeleteVolume = func(ec2Client iaas.IEC2, volumeID *string) error {
		fmt.Printf("Deleting volume: %s\n", *volumeID)
		return nil
	}

	var fakeEC2ClientCreator = func() (iaas.IEC2, error) {
		return &fakeEC2Client{}, nil
	}

	Context("When volumes are provided", func() {
		It("deletes the volumes", func() {
			awsClient, err := iaas.New("AWS", "eu-west-1")
			Expect(err).To(Succeed())
			volumes := []*string{aws.String("volume1"), aws.String("volume2")}
			r, w, _ := os.Pipe()
			tmp := os.Stdout
			defer func() {
				os.Stdout = tmp
			}()
			os.Stdout = w
			go func() {
				err = awsClient.DeleteVolumes(volumes, fakeDeleteVolume, fakeEC2ClientCreator)
				w.Close()
			}()
			stdout, _ := ioutil.ReadAll(r)
			Expect(err).To(Succeed())
			Expect(string(stdout)).To(Equal("Deleting volume: volume1\nDeleting volume: volume2\n"))
		})
	})

	Context("When no volumes are provided", func() {
		It("doesn't delete anything and succeeds", func() {
			awsClient, err := iaas.New("AWS", "eu-west-1")
			Expect(err).To(Succeed())
			volumes := []*string{}
			err = awsClient.DeleteVolumes(volumes, fakeDeleteVolume, fakeEC2ClientCreator)
			Expect(err).To(Succeed())
		})
	})
})
