package config_test

import (
	"fmt"
	"strings"
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
		expectedErr  string
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
			wantErr:     true,
			expectedErr: "--tls-cert requires --tls-key to also be provided",
		},
		{
			name: "TLSKey cannot be set without TLSCert",
			modification: func() DeployArgs {
				args := defaultFields
				args.TLSKey = "a cool key"
				args.Domain = "a cool domain"
				return args
			},
			wantErr:     true,
			expectedErr: "--tls-key requires --tls-cert to also be provided",
		},
		{
			name: "TLSKey and TLSCert require a domain",
			modification: func() DeployArgs {
				args := defaultFields
				args.TLSKey = "a cool key"
				args.TLSCert = "a cool cert"
				return args
			},
			wantErr:     true,
			expectedErr: "custom certificates require --domain to be provided",
		},
		{
			name: "Worker count must be positive",
			modification: func() DeployArgs {
				args := defaultFields
				args.WorkerCount = 0
				return args
			},
			wantErr:     true,
			expectedErr: "minimum number of workers is 1",
		},
		{
			name: "Worker size must be a known value",
			modification: func() DeployArgs {
				args := defaultFields
				args.WorkerSize = "bananas"
				return args
			},
			wantErr:     true,
			expectedErr: fmt.Sprintf("unknown worker size: `bananas`. Valid sizes are: %v", WorkerSizes),
		},
		{
			name: "Web size must be a known value",
			modification: func() DeployArgs {
				args := defaultFields
				args.WebSize = "bananas"
				return args
			},
			wantErr:     true,
			expectedErr: fmt.Sprintf("unknown web node size: `bananas`. Valid sizes are: %v", WebSizes),
		},
		{
			name: "DB size must be a known value",
			modification: func() DeployArgs {
				args := defaultFields
				args.DBSize = "bananas"
				return args
			},
			wantErr:     true,
			expectedErr: fmt.Sprintf("unknown DB size: `bananas`. Valid sizes are:"),
		},
		{
			name: "Github ID requires Github Secret",
			modification: func() DeployArgs {
				args := defaultFields
				args.GithubAuthClientID = "an id"
				return args
			},
			wantErr:     true,
			expectedErr: "--github-auth-client-id requires --github-auth-client-secret to also be provided",
		},
		{
			name: "Github Secret requires Github ID",
			modification: func() DeployArgs {
				args := defaultFields
				args.GithubAuthClientSecret = "super secret"
				return args
			},
			wantErr:     true,
			expectedErr: "--github-auth-client-secret requires --github-auth-client-id to also be provided",
		},
		{
			name: "Tags should be in the format 'key=value'",
			modification: func() DeployArgs {
				args := defaultFields
				args.Tags = []string{"Key=Value", "Cheese=Ham"}
				return args
			},
			wantErr: false,
		},
		{
			name: "Invalid tags should throw a helpful error",
			modification: func() DeployArgs {
				args := defaultFields
				args.Tags = []string{"not a real tag"}
				return args
			},
			wantErr:     true,
			expectedErr: "`not a real tag` is not in the format `key=value`",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.modification()
			err := args.Validate()
			if (err != nil) != tt.wantErr || (err != nil && tt.wantErr && !strings.Contains(err.Error(), tt.expectedErr)) {
				if err != nil {
					t.Errorf("DeployArgs.Validate() %v test failed.\nFailed with error = %v,\nExpected error = %v,\nShould fail %v\nWith args: %#v", tt.name, err.Error(), tt.expectedErr, tt.wantErr, args)
				} else {
					t.Errorf("DeployArgs.Validate() %v test failed.\nShould fail %v\nWith args: %#v", tt.name, tt.wantErr, args)
				}
			}
		})
	}
}
