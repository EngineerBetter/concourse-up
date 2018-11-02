package iaas

import (
	"reflect"
	"testing"
)

func FakeGCPStorage() GCPOption {
	return func(c *GCPProvider) error {
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
	}{
		{
			name: "return aws provider",
			args: args{
				iaas:   "AWS", // it should not matter if it is capitals
				region: "aRegion",
			},
			want:    "AWS",
			wantErr: false,
		}, {
			name: "return gcp provider",
			args: args{
				iaas:   "GCP", // it should not matter if it is capitals
				region: "aRegion",
			},
			want:    "GCP",
			wantErr: false,
		},
		{
			name: "does not care about case",
			args: args{
				iaas:   "aws", // it should not matter if it is capitals
				region: "aRegion",
			},
			want:    "AWS",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.iaas, tt.args.region, FakeGCPStorage())
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
