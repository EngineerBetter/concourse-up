package commands

import (
	"testing"

	"github.com/EngineerBetter/concourse-up/commands/deploy"
)

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
			name: "region should not changed if user provided it",
			args: deploy.Args{
				IAAS:      "AWS",
				AWSRegion: "us-east-1",
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
