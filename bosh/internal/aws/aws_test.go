package aws

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/EngineerBetter/concourse-up/resource"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type mockS3API struct {
	s3iface.S3API
	getObjectOutput *s3.GetObjectOutput
	err             error
}

func (m *mockS3API) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return m.getObjectOutput, m.err
}

func (m *mockS3API) PutObject(in *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return nil, m.err
}

func TestStore_Get(t *testing.T) {
	type fields struct {
		s3     s3iface.S3API
		bucket string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				s3: &mockS3API{
					getObjectOutput: &s3.GetObjectOutput{
						Body: ioutil.NopCloser(strings.NewReader("my object body")),
					},
				},
				bucket: "my bucket",
			},
			args: args{
				key: "state.json",
			},
			want: []byte("my object body"),
		},
		{
			name: "failure",
			fields: fields{
				s3: &mockS3API{
					err: errors.New("an error"),
				},
				bucket: "my bucket",
			},
			args: args{
				key: "state.json",
			},
			wantErr: true,
		},
		{
			name: "not found",
			fields: fields{
				s3: &mockS3API{
					err: awserr.New(s3.ErrCodeNoSuchKey, "no such key", nil),
				},
				bucket: "my bucket",
			},
			args: args{
				key: "state.json",
			},
			wantErr: false,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				s3:     tt.fields.s3,
				bucket: tt.fields.bucket,
			}
			got, err := s.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Store.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_Set(t *testing.T) {
	type fields struct {
		s3     s3iface.S3API
		bucket string
	}
	type args struct {
		key   string
		value []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				s3:     &mockS3API{},
				bucket: "my bucket",
			},
			args: args{
				key: "state.json",
			},
		},
		{
			name: "failure",
			fields: fields{
				s3: &mockS3API{
					err: errors.New("an error"),
				},
				bucket: "my bucket",
			},
			args: args{
				key: "state.json",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				s3:     tt.fields.s3,
				bucket: tt.fields.bucket,
			}
			if err := s.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Store.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvironment_ConfigureDirectorCloudConfig(t *testing.T) {
	type fields struct {
		AZ               string
		PublicSubnetID   string
		PrivateSubnetID  string
		ATCSecurityGroup string
		VMSecurityGroup  string
		Spot             bool
		WorkerType       string
	}
	defaultValidate := func(a, b string) (bool, string) {
		return strings.Contains(a, b), fmt.Sprintf("Environment.ConfigureDirectorCloudConfig()\nexpected '%v'\nreceived '%v'", a, b)
	}
	tests := []struct {
		name        string
		fields      fields
		cloudConfig string
		want        string
		wantErr     bool
		validate    func(string, string) (bool, string)
	}{
		{
			name: "Success- template rendered",
			fields: fields{
				AZ:               "eu-west-1",
				PublicSubnetID:   "12345",
				PrivateSubnetID:  "67890",
				ATCSecurityGroup: "00000",
				WorkerType:       "m4",
			},
			cloudConfig: "availability_zone: {{ .AvailabilityZone }}\n public_subnet_id: {{ .PublicSubnetID }}\n private_subnet_id: {{ .PrivateSubnetID }}\n atc_security_group: {{ .ATCSecurityGroupID }}\n spot: {{ .Spot }}\nworker_type: {{ .WorkerType}}",
			want:        "availability_zone: eu-west-1\n public_subnet_id: 12345\n private_subnet_id: 67890\n atc_security_group: 00000\n spot: false\nworker_type: m4",
			wantErr:     false,
			validate:    defaultValidate,
		},
		{
			name: "Success- spot instance rendered",
			fields: fields{
				Spot: true,
			},
			cloudConfig: resource.AWSDirectorCloudConfig,
			want:        "spot_bid_price",
			wantErr:     false,
			validate: func(a, b string) (bool, string) {
				re := regexp.MustCompile(b)
				n := re.FindAllString(a, -1)
				return len(n) == 10, fmt.Sprintf("Expected 10 appearances of '%v' in\n'%+v'\ninstead got %+v", b, a, len(n))
			},
		},

		{
			name: "Success- running with no spot",
			fields: fields{
				Spot: false,
			},
			cloudConfig: resource.AWSDirectorCloudConfig,
			want:        "spot_bid_price",
			wantErr:     false,
			validate: func(a, b string) (bool, string) {
				re := regexp.MustCompile(b)
				n := re.FindAllString(a, -1)
				return len(n) == 0, fmt.Sprintf("Expected 0 appearances of '%v' in\n'%+v'\ninstead got %+v", b, a, len(n))
			},
		},
		{
			name: "Success- worker type is m5",
			fields: fields{
				WorkerType: "m5",
			},
			cloudConfig: resource.AWSDirectorCloudConfig,
			want:        "instance_type: m5",
			wantErr:     false,
			validate: func(a, b string) (bool, string) {
				re := regexp.MustCompile(b)
				n := re.FindAllString(a, -1)
				return len(n) == 7, fmt.Sprintf("Expected 7 appearances of '%v' in\n'%+v'\ninstead got %+v", b, a, len(n))
			},
		},
		{
			name: "Success- m4 worker type is m4",
			fields: fields{
				WorkerType: "m4",
			},
			cloudConfig: resource.AWSDirectorCloudConfig,
			want:        "instance_type: m4",
			wantErr:     false,
			validate: func(a, b string) (bool, string) {
				re := regexp.MustCompile(b)
				n := re.FindAllString(a, -1)
				return len(n) == 7, fmt.Sprintf("Expected 7 appearances of '%v' in\n'%+v'\ninstead got %+v", b, a, len(n))
			},
		},
		{
			name:        "Failure- template not rendered",
			cloudConfig: "non_existant_key: {{ .NonExistantKey }}",
			want:        "",
			wantErr:     true,
			validate:    defaultValidate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Environment{
				AZ:               tt.fields.AZ,
				PublicSubnetID:   tt.fields.PublicSubnetID,
				PrivateSubnetID:  tt.fields.PrivateSubnetID,
				ATCSecurityGroup: tt.fields.ATCSecurityGroup,
				VMSecurityGroup:  tt.fields.VMSecurityGroup,
				Spot:             tt.fields.Spot,
				WorkerType:       tt.fields.WorkerType,
			}
			got, err := e.ConfigureDirectorCloudConfig(tt.cloudConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("Environment.ConfigureDirectorCloudConfig()\nerror expected:  %v\nreceived error:  %v", tt.wantErr, err)
				return
			}
			passed, message := tt.validate(got, tt.want)
			if !passed {
				t.Errorf(message)
			}
		})
	}
}

func TestEnvironment_ConfigureConcourseStemcell(t *testing.T) {
	type args struct {
		versions string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "parse versions and provide a valid stemcell url",
			args: args{
				versions: `[{
					"type": "replace",
					"path": "/stemcells/alias=xenial/version",
					"value": "97.19"
				  }]`,
			},
			want:    fmt.Sprintf("https://s3.amazonaws.com/bosh-aws-light-stemcells/light-bosh-stemcell-%s-aws-xen-hvm-ubuntu-xenial-go_agent.tgz", "97.19"),
			wantErr: false,
		},
		{
			name: "parse versions and fail if stemcell is not xenial",
			args: args{
				versions: `[{
					"type": "replace",
					"path": "/stemcells/alias=zenial/version",
					"value": "97.19"
				  }]`,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fail if no xenial stemcell exists",
			args: args{
				versions: `[]`,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Environment{}
			got, err := e.ConfigureConcourseStemcell(tt.args.versions)
			if (err != nil) != tt.wantErr {
				t.Errorf("Environment.ConfigureConcourseStemcell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Environment.ConfigureConcourseStemcell() = %v, want %v", got, tt.want)
			}
		})
	}
}
