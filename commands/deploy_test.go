package commands

import (
	"testing"

	"github.com/EngineerBetter/concourse-up/commands/deploy"
)

// when the equivalent of concourse-up --iaas gcp deploy is called
// deplayAction(c, fakeIaasFactory, fakeConcourseClient, fakeTfClient)
// concourse.NewClient is invoked, passing a configClient with a Region of "europe-west1"

func Test_setZoneAndRegion(t *testing.T) {
	tests := []struct {
		name           string
		args           deploy.DeployArgs
		wantErr        bool
		expectedRegion string
	}{
		{
			name: "success - region defaults to europe-west1 when iaas is gcp",
			args: deploy.DeployArgs{
				IAAS: "GCP",
			},
			expectedRegion: "europe-west1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := setZoneAndRegion(deploy.DeployArgs{})

			if err == nil && tt.wantErr {
				t.Errorf("setZoneAndRegion() error = %v, wantErr %v", err, tt.wantErr)
			}

			if actual.AWSRegion != tt.expectedRegion {
				t.Errorf("setZoneAndRegion() region = %v, expected %v", actual.AWSRegion, tt.expectedRegion)
			}
		})
	}
}
