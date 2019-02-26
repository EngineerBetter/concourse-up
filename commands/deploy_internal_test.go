package commands

import (
	"fmt"
	"testing"

	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/testsupport"

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
		providerRegion string
		wantErr        bool
		expectedRegion string
	}{
		{
			name: "region should default to eu-west-1 when iaas is AWS",
			args: deploy.Args{
				IAAS: "AWS",
			},
			providerRegion: "eu-west-1",
			expectedRegion: "eu-west-1",
		},
		{
			name: "region should default to europe-west1 when iaas is GCP",
			args: deploy.Args{
				IAAS: "GCP",
			},
			providerRegion: "europe-west1",
			expectedRegion: "europe-west1",
		},
		{
			name: "region should change if user provided it",
			args: deploy.Args{
				IAAS:        "AWS",
				Region:      "us-east-1",
				RegionIsSet: true,
			},
			providerRegion: "eu-west-1",
			expectedRegion: "us-east-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := setZoneAndRegion(tt.providerRegion, tt.args)

			if err == nil && tt.wantErr {
				t.Errorf("setZoneAndRegion() error = %v, wantErr %v", err, tt.wantErr)
			}

			if actual.Region != tt.expectedRegion {
				t.Errorf("setZoneAndRegion() region = %v, expected %v", actual.Region, tt.expectedRegion)
			}
		})
	}
}

func Test_validateNameLength(t *testing.T) {
	type args struct {
		name         string
		providerName iaas.Name
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "It is GCP with a valid name length",
			args: args{
				name:         "a-name",
				providerName: iaas.GCP,
			},
			wantErr: false,
		},
		{
			name: "It is GCP with a invalid name length",
			args: args{
				name:         "a-name-that-is-long-enough-make-this-fail",
				providerName: iaas.GCP,
			},
			wantErr: true,
		},
		{
			name: "It is GCP with an invalid name length",
			args: args{
				name:         "a-name",
				providerName: iaas.GCP,
			},
			wantErr: false,
		},
		{
			name: "It is GCP with an invalid name length",
			args: args{
				name:         "a-name-that-is-long-enough-make-this-fail-on-gcp",
				providerName: iaas.GCP,
			},
			wantErr: true,
		},
		{
			name: "It is AWS with a valid name length",
			args: args{
				name:         "a-name",
				providerName: iaas.AWS,
			},
			wantErr: false,
		},
		{
			name: "It is AWS with an invalid name length",
			args: args{
				name:         "a-name-that-is-long-enough-make-this-fail-on-gcp",
				providerName: iaas.AWS,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateNameLength(tt.args.name, tt.args.providerName); (err != nil) != tt.wantErr {
				t.Errorf("validateNameLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateCidrRanges(t *testing.T) {
	testsupport.SetupFakeCredsForGCPProvider(t)
	gcpProvider, err := iaas.New(iaas.GCP, "europe-west1")
	if err != nil {
		t.Fatalf("Error creating GCP provider in test: [%v]", err)
	}
	awsProvider, err := iaas.New(iaas.AWS, "eu-west-1")
	if err != nil {
		t.Fatalf("Error creating AWS provider in test: [%v]", err)
	}

	type args struct {
		provider    iaas.Provider
		networkCidr string
		publicCidr  string
		privateCidr string
		rds1Cidr    string
		rds2Cidr    string
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		desiredErrMsg string
	}{
		{
			name: "does not err if no flags provided",
			args: args{
				provider: awsProvider,
			},
			wantErr: false,
		}, //
		{
			name: "errs if public range is provided and private is not",
			args: args{
				provider:    gcpProvider,
				networkCidr: "10.0.0.0/16",
				publicCidr:  "10.0.0.0/24",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - both public-subnet-range and private-subnet-range must be provided",
		},
		{
			name: "errs if private range is provided and public is not",
			args: args{
				provider:    gcpProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - both public-subnet-range and private-subnet-range must be provided",
		},
		{
			name: "errs if provider is AWS and default range is not provided",
			args: args{
				provider:    awsProvider,
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.1.0/24",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - vpc-network-range must be provided when using AWS",
		},
		{
			name: "errs if provider is AWS and rds1 range is not provided",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.1.0/24",
				rds2Cidr:    "10.0.4.0/24",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - both rds1-subnet-range and rds2-subnet-range must be provided",
		},
		{
			name: "errs if provider is AWS and rds2 range is not provided",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.1.0/24",
				rds1Cidr:    "10.0.2.0/24",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - both rds1-subnet-range and rds2-subnet-range must be provided",
		},
		{
			name: "errs if default range isn't a CIDR",
			args: args{
				provider:    awsProvider,
				networkCidr: "aNetwork",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.1.0/24",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - vpc-network-range is not a valid CIDR",
		},
		{
			name: "errs if public range isn't a CIDR",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "aPublicNetwork",
				rds1Cidr:    "10.0.2.0/16",
				rds2Cidr:    "10.0.3.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - public-subnet-range is not a valid CIDR",
		},
		{
			name: "errs if private range isn't a CIDR",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "aPrivateNetwork",
				publicCidr:  "10.0.1.0/24",
				rds1Cidr:    "10.0.2.0/16",
				rds2Cidr:    "10.0.3.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - private-subnet-range is not a valid CIDR",
		},
		{
			name: "errs if rds1 range isn't a CIDR",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.1.0/24",
				rds1Cidr:    "aPrivateNetwork",
				rds2Cidr:    "10.0.3.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - rds1-subnet-range is not a valid CIDR",
		},
		{
			name: "errs if rds2 range isn't a CIDR",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.1.0/24",
				rds1Cidr:    "10.0.2.0/16",
				rds2Cidr:    "aPrivateNetwork",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - rds2-subnet-range is not a valid CIDR",
		},
		{
			name: "errs if provider is AWS and public range is not in default range",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "172.0.0.0/24",
				rds1Cidr:    "10.0.2.0/16",
				rds2Cidr:    "10.0.3.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - public-subnet-range must be within vpc-network-range",
		},
		{
			name: "errs if provider is AWS and private range is not in default range",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "172.0.0.0/24",
				publicCidr:  "10.0.1.0/24",
				rds1Cidr:    "10.0.2.0/16",
				rds2Cidr:    "10.0.3.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - private-subnet-range must be within vpc-network-range",
		},
		{
			name: "errs if provider is AWS and rds1 range is not in default range",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.1.0/24",
				rds1Cidr:    "192.168.2.0/16",
				rds2Cidr:    "10.0.3.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - rds1-subnet-range must be within vpc-network-range",
		},
		{
			name: "errs if provider is AWS and rds2 range is not in default range",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.1.0/24",
				rds1Cidr:    "10.0.3.0/16",
				rds2Cidr:    "192.168.2.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - rds2-subnet-range must be within vpc-network-range",
		},
		{
			name: "errs if public range overlaps with private range",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.0.0/24",
				rds1Cidr:    "10.0.3.0/16",
				rds2Cidr:    "10.0.4.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - public-subnet-range must not overlap private-network-range",
		},
		{
			name: "errs if network cidr range is not big enough (16 usable IPs)",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/28",
				privateCidr: "10.0.0.0/32",
				publicCidr:  "10.0.0.0/24",
				rds1Cidr:    "10.0.3.0/16",
				rds2Cidr:    "10.0.4.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - vpc-network-range is not big enough, at least /26 needed.",
		},
		{
			name: "errs if private cidr range is not big enough (8 usable IPs)",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/32",
				publicCidr:  "10.0.0.0/24",
				rds1Cidr:    "10.0.3.0/16",
				rds2Cidr:    "10.0.4.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - private-subnet-range is not big enough, at least /28 needed.",
		},
		{
			name: "errs if public cidr range is not big enough  (8 usable IPs)",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.0.0/32",
				rds1Cidr:    "10.0.3.0/16",
				rds2Cidr:    "10.0.4.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - public-subnet-range is not big enough, at least /28 needed.",
		},
		{
			name: "errs if rds1 cidr range is not big enough  (8 usable IPs)",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.0.0/24",
				rds1Cidr:    "10.0.3.0/32",
				rds2Cidr:    "10.0.4.0/16",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - rds1-subnet-range is not big enough, at least /29 needed.",
		},
		{
			name: "errs if rds2 cidr range is not big enough  (8 usable IPs)",
			args: args{
				provider:    awsProvider,
				networkCidr: "10.0.0.0/16",
				privateCidr: "10.0.0.0/24",
				publicCidr:  "10.0.0.0/24",
				rds1Cidr:    "10.0.3.0/24",
				rds2Cidr:    "10.0.4.0/32",
			},
			wantErr:       true,
			desiredErrMsg: "error validating CIDR ranges - rds2-subnet-range is not big enough, at least /29 needed.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCidrRanges(tt.args.provider, tt.args.networkCidr, tt.args.publicCidr, tt.args.privateCidr, tt.args.rds1Cidr, tt.args.rds2Cidr)

			if (err == nil && tt.wantErr) || (err != nil && !tt.wantErr) {
				t.Errorf("validateCidrRanges() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				if err.Error() != tt.desiredErrMsg {
					t.Errorf("validateCidrRanges() error message = [%v], desiredErrMsg [%v]", err.Error(), tt.desiredErrMsg)
				}
			}
		})
	}
}
