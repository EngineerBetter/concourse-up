package config_test

import (
	"testing"

	. "github.com/EngineerBetter/concourse-up/config"
)

func TestDeployArgs_Validate(t *testing.T) {
	defaultFields := DeployArgs{
		AllowIPs:               "0.0.0.0",
		AWSRegion:              "eu-west-1",
		DBSize:                 "small",
		DBSizeIsSet:            false,
		Domain:                 "",
		GithubAuthClientID:     "",
		GithubAuthClientSecret: "",
		IAAS:        "AWS",
		SelfUpdate:  false,
		TLSCert:     "",
		TLSKey:      "",
		WebSize:     "small",
		WorkerCount: 1,
		WorkerSize:  "xlarge",
	}
	tests := []struct {
		name         string
		modification func() DeployArgs
		wantErr      bool
	}{
		{
			name: "Default args",
			modification: func() DeployArgs {
				return defaultFields
			},
			wantErr: false,
		},
		{
			name: "All cert fields should be set",
			modification: func() DeployArgs {
				args := defaultFields
				args.TLSCert = "a cool cert"
				args.TLSKey = "a cool key"
				args.Domain = "a cool domain"
				return args
			},
			wantErr: false,
		},
		{
			name: "TLSCert cannot be set without TLSKey",
			modification: func() DeployArgs {
				args := defaultFields
				args.TLSCert = "a cool cert"
				args.Domain = "a cool domain"
				return args
			},
			wantErr: true,
		},
		{
			name: "TLSKey cannot be set without TLSCert",
			modification: func() DeployArgs {
				args := defaultFields
				args.TLSKey = "a cool key"
				args.Domain = "a cool domain"
				return args
			},
			wantErr: true,
		},
		{
			name: "TLSKey and TLSCert require a domain",
			modification: func() DeployArgs {
				args := defaultFields
				args.TLSKey = "a cool key"
				args.TLSCert = "a cool cert"
				return args
			},
			wantErr: true,
		},
		{
			name: "Worker count must be positive",
			modification: func() DeployArgs {
				args := defaultFields
				args.WorkerCount = 0
				return args
			},
			wantErr: true,
		},
		{
			name: "Worker size must be a known value",
			modification: func() DeployArgs {
				args := defaultFields
				args.WorkerSize = "bananas"
				return args
			},
			wantErr: true,
		},
		{
			name: "Web size must be a known value",
			modification: func() DeployArgs {
				args := defaultFields
				args.WebSize = "bananas"
				return args
			},
			wantErr: true,
		},
		{
			name: "DB size must be a known value",
			modification: func() DeployArgs {
				args := defaultFields
				args.DBSize = "bananas"
				return args
			},
			wantErr: true,
		},
		{
			name: "Github ID requires Github Secret",
			modification: func() DeployArgs {
				args := defaultFields
				args.GithubAuthClientID = "an id"
				return args
			},
			wantErr: true,
		},
		{
			name: "Github Secret requires Github ID",
			modification: func() DeployArgs {
				args := defaultFields
				args.GithubAuthClientSecret = "super secret"
				return args
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.modification()
			if err := args.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("DeployArgs.Validate() %v test failed.\nGot error = %v,\nExpected %v\nWith args: %#v", tt.name, err, tt.wantErr, args)
			}
		})
	}
}
