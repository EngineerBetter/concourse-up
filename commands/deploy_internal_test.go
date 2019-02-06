package commands

import (
	"fmt"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/testsupport"
	"testing"

	"github.com/EngineerBetter/concourse-up/commands/deploy"
)

func Test_regionFromZone(t *testing.T) {
	type args struct {
		zone string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name:  "a valid zone returns a valid region",
			args:  args{zone: "eu-west-1b"},
			want:  "eu-west-1",
			want1: fmt.Sprintf("No region provided, please note that your zone will be paired with a matching region.\nThis region: %s is used for deployment.\n", "eu-west-1"),
		},
		{
			name:  "an invalid zone returns empty region",
			args:  args{zone: "wrong-zone"},
			want:  "",
			want1: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := regionFromZone(tt.args.zone)
			if got != tt.want {
				t.Errorf("regionFromZone() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("regionFromZone() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
func Test_zoneBelongsToRegion(t *testing.T) {
	type args struct {
		zone   string
		region string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "zone belongs to region",
			args: args{
				zone:   "us-east1c",
				region: "us-east1",
			},
			wantErr: false,
		},
		{
			name: "zone does not belong to region",
			args: args{
				zone:   "us-east1a",
				region: "eu-east-1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := zoneBelongsToRegion(tt.args.zone, tt.args.region); (err != nil) != tt.wantErr {
				t.Errorf("zoneBelongsToRegion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setZoneAndRegion(t *testing.T) {
	tests := []struct {
		name           string
		args           deploy.Args
		wantErr        bool
		expectedRegion string
	}{
		{
			name: "region should default to eu-west-1 when iaas is AWS",
			args: deploy.Args{
				IAAS: "AWS",
			},
			expectedRegion: "eu-west-1",
		},
		{
			name: "region should default to europe-west1 when iaas is GCP",
			args: deploy.Args{
				IAAS: "GCP",
			},
			expectedRegion: "europe-west1",
		},
		{
			name: "region should not change if user provided it",
			args: deploy.Args{
				IAAS:           "AWS",
				AWSRegion:      "us-east-1",
				AWSRegionIsSet: true,
			},
			expectedRegion: "us-east-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := setZoneAndRegion(tt.args)

			if err == nil && tt.wantErr {
				t.Errorf("setZoneAndRegion() error = %v, wantErr %v", err, tt.wantErr)
			}

			if actual.AWSRegion != tt.expectedRegion {
				t.Errorf("setZoneAndRegion() region = %v, expected %v", actual.AWSRegion, tt.expectedRegion)
			}
		})
	}
}

func Test_validateNameLength(t *testing.T) {
	testsupport.SetupFakeCredsForGCPProvider(t)
	gcpProvider, err := iaas.New("GCP", "europe-west1")
	if err != nil {
		t.Fatalf("Error initialisting iaas.Provider: [%v]", err)
	}

	awsProvider, err := iaas.New("AWS", "eu-west-1")
	if err != nil {
		t.Fatalf("Error initialisting iaas.Provider: [%v]", err)
	}

	type args struct {
		name     string
		provider iaas.Provider
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "It is GCP with a valid name length",
			args: args{
				name:     "a-name",
				provider: gcpProvider,
			},
			wantErr: false,
		},
		{
			name: "It is GCP with a invalid name length",
			args: args{
				name:     "a-name-that-is-long-enough-make-this-fail",
				provider: gcpProvider,
			},
			wantErr: true,
		},
		{
			name: "It is gcP with an invalid name length",
			args: args{
				name:     "a-name",
				provider: gcpProvider,
			},
			wantErr: false,
		},
		{
			name: "It is gcP with an invalid name length",
			args: args{
				name:     "a-name-that-is-long-enough-make-this-fail-on-gcp",
				provider: gcpProvider,
			},
			wantErr: true,
		},
		{
			name: "It is AWS with a valid name length",
			args: args{
				name:     "a-name",
				provider: awsProvider,
			},
			wantErr: false,
		},
		{
			name: "It is AWS with an invalid name length",
			args: args{
				name:     "a-name-that-is-long-enough-make-this-fail-on-gcp",
				provider: awsProvider,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateNameLength(tt.args.name, tt.args.provider); (err != nil) != tt.wantErr {
				t.Errorf("validateNameLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
