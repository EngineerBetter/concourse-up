package commands

import (
	"fmt"
	"testing"
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
