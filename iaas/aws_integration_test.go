// +build integration

package iaas

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestAWSProvider_DeleteVolumes(t *testing.T) {
	region := "us-east-1"
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	svc := ec2.New(sess)
	input := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String("us-east-1a"),
		Size:             aws.Int64(1),
		VolumeType:       aws.String("gp2"),
	}
	result, err := svc.CreateVolume(input)
	if err != nil {
		t.Error("Cannot create a volume")
	}

	type args struct {
		volumes      []string
		deleteVolume func(ec2Client IEC2, volumeID *string) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "find all created volumes for deletetion",
			args: args{
				volumes: []string{*result.VolumeId},
				deleteVolume: func(ec2Client IEC2, volumeID *string) error {
					fmt.Printf("volume %s is about to be deleted\n", *volumeID)
					return nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			a := &AWSProvider{
				sess: sess,
			}

			if err := a.DeleteVolumes(tt.args.volumes, DeleteVolume); (err != nil) != tt.wantErr {
				t.Errorf("AWSProvider.DeleteVolumes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAWSProvider_Zone(t *testing.T) {
	type fields struct {
		region     string
		workerType string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "return a valid zone for m5.large",
			fields: fields{
				region:     "us-east-1",
				workerType: "m5.large",
			},
			want: "us-east-1a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess, _ := session.NewSession(&aws.Config{
				Region: aws.String(tt.fields.region),
			})
			a := &AWSProvider{
				sess:       sess,
				workerType: tt.fields.workerType,
			}
			if got := a.Zone(); got != tt.want {
				t.Errorf("AWSProvider.Zone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAWSProvider_listZones(t *testing.T) {
	tests := []struct {
		region  string
		name    string
		want    []string
		wantErr bool
	}{
		{
			region:  "us-east-1",
			name:    "provide some zones with no error",
			want:    []string{"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1e", "us-east-1f"},
			wantErr: false,
		},
		{
			region:  "us-somewhere-1",
			name:    "error out on unknown region",
			want:    []string{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess, _ := session.NewSession(&aws.Config{
				Region: aws.String(tt.region),
			})
			a := &AWSProvider{
				sess: sess,
			}
			got, err := a.listZones()
			if (err != nil) != tt.wantErr {
				t.Errorf("AWSProvider.listZones() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AWSProvider.listZones() = %v, want %v", got, tt.want)
			}
		})
	}
}
