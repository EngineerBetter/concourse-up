package config

import (
	"testing"
)

func Test_generateDefaultConfig(t *testing.T) {
	type args struct {
		iaas         string
		project      string
		deployment   string
		configBucket string
		region       string
		namespace    string
	}
	tests := []struct {
		name    string
		args    args
		isValid func(Config) (bool, string)
		wantErr bool
	}{
		{
			name: "providing the namespace",
			args: args{
				iaas:         "iass",
				project:      "mySHINYBRANDnewProject",
				deployment:   "myDeployment",
				configBucket: "some/bucket",
				region:       "in-the-west",
				namespace:    "deep",
			},
			isValid: func(c Config) (bool, string) {
				return c.Namespace == "deep", c.Namespace + " != " + "deep"
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateDefaultConfig(tt.args.iaas, tt.args.project, tt.args.deployment, tt.args.configBucket, tt.args.region, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateDefaultConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if passed, outcome := tt.isValid(got); !passed {
				t.Errorf("generateDefaultConfig() outcome %v", outcome)
				return
			}
		})
	}
}
