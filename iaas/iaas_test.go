package iaas_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/testsupport"
)

func FakeGCPStorage() iaas.GCPOption {
	return func(c *iaas.GCPProvider) error {
		return nil
	}
}
func TestNew(t *testing.T) {
	type args struct {
		iaas    iaas.Name
		region  string
		project string
	}

	tests := []struct {
		name    string
		args    args
		want    iaas.Name
		wantErr bool
		setup   func(t *testing.T) string
		cleanup func(t *testing.T, s string)
	}{
		{
			name: "return aws provider",
			args: args{
				iaas:   iaas.AWS, // it should not matter if it is capitals
				region: "aRegion",
			},
			want:    iaas.AWS,
			wantErr: false,
			setup:   func(t *testing.T) string { return "" },
			cleanup: func(t *testing.T, s string) {},
		}, {
			name: "return gcp provider",
			args: args{
				iaas:   iaas.GCP, // it should not matter if it is capitals
				region: "aRegion",
			},
			want:    iaas.GCP,
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
				iaas:   iaas.AWS, // it should not matter if it is capitals
				region: "aRegion",
			},
			want:    iaas.AWS,
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

func TestAssosiate(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    iaas.Name
		wantErr bool
	}{
		{
			name:    "get the GCP Name successfully",
			arg:     "GCP",
			want:    iaas.GCP,
			wantErr: false,
		},
		{
			name:    "get the GCP Name successfully case insensitive",
			arg:     "GcP",
			want:    iaas.GCP,
			wantErr: false,
		},
		{
			name:    "get the AWS Name successfully",
			arg:     "AWS",
			want:    iaas.AWS,
			wantErr: false,
		},
		{
			name:    "fail on unknown iaas name",
			arg:     "aProvider",
			want:    iaas.Unknown,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := iaas.Assosiate(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Assosiate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Assosiate() = %v, want %v", got, tt.want)
			}
		})
	}
}
