package iaas_test

import (
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/testsupport"
	"os"
	"reflect"
	"testing"
)

func FakeGCPStorage() iaas.GCPOption {
	return func(c *iaas.GCPProvider) error {
		return nil
	}
}
func TestNew(t *testing.T) {
	type args struct {
		iaas    string
		region  string
		project string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
		setup   func(t *testing.T) string
		cleanup func(t *testing.T, s string)
	}{
		{
			name: "return aws provider",
			args: args{
				iaas:   "AWS", // it should not matter if it is capitals
				region: "aRegion",
			},
			want:    "AWS",
			wantErr: false,
			setup:   func(t *testing.T) string { return "" },
			cleanup: func(t *testing.T, s string) {},
		}, {
			name: "return gcp provider",
			args: args{
				iaas:   "GCP", // it should not matter if it is capitals
				region: "aRegion",
			},
			want:    "GCP",
			wantErr: false,
			setup: func(t *testing.T) string {
				return testsupport.SetupFakeCredsForGCPProvider(t)
			},
			cleanup: func(t *testing.T, s string) {
				if err := os.Remove(s); err != nil {
					t.Error("Could not delete GCP credentials file")
				}
			},
		},
		{
			name: "does not care about case",
			args: args{
				iaas:   "aws", // it should not matter if it is capitals
				region: "aRegion",
			},
			want:    "AWS",
			wantErr: false,
			setup:   func(t *testing.T) string { return "" },
			cleanup: func(t *testing.T, s string) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := tt.setup(t)
			got, err := iaas.New(tt.args.iaas, tt.args.region)
			tt.cleanup(t, tmp)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.IAAS(), tt.want) {
				t.Errorf("New() = %v, want %v", got.IAAS(), tt.want)
			}
		})
	}
}
