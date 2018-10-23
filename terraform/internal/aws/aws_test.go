package aws_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/EngineerBetter/concourse-up/terraform/internal/aws"
)

func TestInputVars_ConfigureTerraform(t *testing.T) {
	type FakeInputVars struct {
		Deployment     string
		Project        string
		Region         string
		SourceAccessIP string
		MultiAZRDS     bool
	}
	type args struct {
		terraformContents string
	}
	tests := []struct {
		name          string
		fakeInputVars FakeInputVars
		args          string
		want          string
		wantErr       bool
	}{
		{name: "Success",
			fakeInputVars: FakeInputVars{
				Deployment:     "fakeDeployment",
				Project:        "fakeProject",
				Region:         "eu-west-1",
				SourceAccessIP: "fakeSourceIP",
				MultiAZRDS:     true,
			},
			args:    "<% .Region %>\n <% .Deployment %>\n <% .Project %>\n <% .SourceAccessIP %>\n <% .MultiAZRDS %>\n",
			want:    "eu-west-1\n fakeDeployment\n fakeProject\n fakeSourceIP\n true\n",
			wantErr: false,
		},
		{name: "Failure",
			fakeInputVars: FakeInputVars{},
			args:          "<% .FakeKey %> \n",
			want:          "",
			wantErr:       true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v := &aws.InputVars{
				Deployment:     test.fakeInputVars.Deployment,
				Project:        test.fakeInputVars.Project,
				Region:         test.fakeInputVars.Region,
				SourceAccessIP: test.fakeInputVars.SourceAccessIP,
				MultiAZRDS:     test.fakeInputVars.MultiAZRDS,
			}
			got, err := v.ConfigureTerraform(test.args)
			if (err != nil) != test.wantErr {
				t.Errorf("InputVars.ConfigureTerraform() test case \"%s\" failed\nReturned error %v\nExpects an error: %v", test.name, err, test.wantErr)
				return
			}
			if got != test.want {
				t.Errorf("InputVars.ConfigureTerraform() test case \"%s\" failed\nReturned value \"%v\"\nExpected value \"%v\"", test.name, got, test.want)
			}
		})
	}
}

func TestMetadata_Get(t *testing.T) {
	type fields struct {
		VPCID aws.MetadataStringValue
	}
	tests := []struct {
		name    string
		fields  fields
		fakeKey string
		want    string
		wantErr bool
	}{{
		name: "Success",
		fields: fields{
			VPCID: aws.MetadataStringValue{
				Value: "fakeMetadataStringValue",
			},
		},
		want:    "fakeMetadataStringValue",
		fakeKey: "VPCID",
	},
		{
			name: "Failure",
			fields: fields{
				VPCID: aws.MetadataStringValue{
					Value: "fakeMetadataStringValue",
				},
			},
			want:    "",
			fakeKey: "FakeKey",
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata := &aws.Metadata{
				VPCID: test.fields.VPCID,
			}
			got, err := metadata.Get(test.fakeKey)
			if (err != nil) != test.wantErr {
				t.Errorf("Metadata.Get() test case  \"%s\" failed\nReturned error %v\nExpects an error: %v", test.name, err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("Metadata.Get() test case \"%s\" failed\nReturned value \"%v\"\nExpected value \"%v\"", test.name, got, test.want)
			}
		})
	}
}

func TestInputVars_Build(t *testing.T) {
	tests := []struct {
		fakeInputVars aws.InputVars
		name          string
		data          map[string]interface{}
		wantErr       bool
	}{
		{
			name: "Success",
			data: map[string]interface{}{
				"AllowIPs":               "allowips",
				"AvailabilityZone":       "availabilityzone",
				"ConfigBucket":           "configbucket",
				"Deployment":             "deployment",
				"HostedZoneID":           "hostedzoneid",
				"HostedZoneRecordPrefix": "hostedzonerecordprefix",
				"Namespace":              "namespace",
				"Project":                "project",
				"PublicKey":              "publickey",
				"RDSDefaultDatabaseName": "rdsdefaultdatabasename",
				"RDSInstanceClass":       "rdsinstanceclass",
				"RDSPassword":            "rdspassword",
				"RDSUsername":            "rdsusername",
				"Region":                 "region",
				"SourceAccessIP":         "sourceaccessip",
				"TFStatePath":            "tfstatepath",
				"MultiAZRDS":             true,
			},
			fakeInputVars: aws.InputVars{},
		},
		{
			name: "Failure",
			data: map[string]interface{}{
				"AvailabilityZone":       "availabilityzone",
				"ConfigBucket":           "configbucket",
				"Deployment":             "deployment",
				"HostedZoneID":           "hostedzoneid",
				"HostedZoneRecordPrefix": "hostedzonerecordprefix",
				"Namespace":              "namespace",
				"Project":                "project",
				"PublicKey":              "publickey",
				"RDSDefaultDatabaseName": "rdsdefaultdatabasename",
				"RDSInstanceClass":       "rdsinstanceclass",
				"RDSPassword":            "rdspassword",
				"RDSUsername":            "rdsusername",
				"Region":                 "region",
				"SourceAccessIP":         "sourceaccessip",
				"TFStatePath":            "tfstatepath",
				"MultiAZRDS":             true,
			},
			wantErr:       true,
			fakeInputVars: aws.InputVars{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.fakeInputVars.Build(test.data); (err != nil) != test.wantErr {
				t.Errorf("InputVars.Build() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestMetadata_Init(t *testing.T) {
	tests := []struct {
		name          string
		buffer        *bytes.Buffer
		data          string
		keyToSet      string
		expectedValue string
		wantErr       bool
	}{
		{
			name:          "Success",
			data:          `{"atc_public_ip":{"sensitive":false,"type": "string","value": "fakeIP"}}`,
			keyToSet:      "ATCPublicIP",
			expectedValue: "fakeIP",
		},
		{
			name:          "Failure",
			keyToSet:      "ATCPublicIP",
			expectedValue: "",
			wantErr:       true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata := &aws.Metadata{}
			buffer := bytes.NewBuffer(nil)
			buffer.WriteString(test.data)
			if err := metadata.Init(buffer); (err != nil) != test.wantErr {
				t.Errorf("Metadata.Init() error = %v, wantErr %v", err, test.wantErr)
			}
			mm := reflect.ValueOf(metadata)
			m := mm.Elem()
			mv := m.FieldByName(test.keyToSet)
			value := mv.FieldByName("Value").String()
			if value != test.expectedValue {
				t.Errorf("Metadata.Init() test case %s\nfailed testing key %s\nexpected value %s\nreceived value %s\n", test.name, test.keyToSet, value, test.expectedValue)
			}

		})
	}
}
