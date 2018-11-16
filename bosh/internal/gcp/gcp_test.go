package gcp

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

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
		Zone              string
		PublicSubnetwork  string
		PrivateSubnetwork string
		Preemptible       bool
	}
	tests := []struct {
		name        string
		fields      fields
		cloudConfig string
		want        string
		wantErr     bool
	}{
		{
			name: "Success- template rendered",
			fields: fields{
				Zone:              "europe-west1-b",
				PublicSubnetwork:  "12345",
				PrivateSubnetwork: "67890",
				Preemptible:       true,
			},
			cloudConfig: "zone: {{ .Zone }}\n public_subnet_id: {{ .PublicSubnetwork }}\n private_subnet_id: {{ .PrivateSubnetwork }}\n preemptible: {{ .Preemptible }}",
			want:        "zone: europe-west1-b\n public_subnet_id: 12345\n private_subnet_id: 67890\n preemptible: true",
			wantErr:     false,
		},
		{
			name:        "Failure- template not rendered",
			cloudConfig: "non_existant_key: {{ .NonExistantKey }}",
			want:        "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Environment{
				Zone:              tt.fields.Zone,
				PublicSubnetwork:  tt.fields.PublicSubnetwork,
				PrivateSubnetwork: tt.fields.PrivateSubnetwork,
				Preemptible:       tt.fields.Preemptible,
			}
			got, err := e.ConfigureDirectorCloudConfig(tt.cloudConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("Environment.ConfigureDirectorCloudConfig()\nerror expected:  %v\nreceived error:  %v", tt.wantErr, err)
				return
			}
			if got != tt.want {
				t.Errorf("Environment.ConfigureDirectorCloudConfig()\nexpected %v\nreceived %v", tt.want, got)
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
			want:    fmt.Sprintf("https://s3.amazonaws.com/bosh-gce-light-stemcells/light-bosh-stemcell-%s-google-kvm-ubuntu-xenial-go_agent.tgz", "97.19"),
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
