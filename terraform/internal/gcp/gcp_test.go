package gcp_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/EngineerBetter/concourse-up/terraform/internal/gcp"
)

func TestInputVars_ConfigureTerraform(t *testing.T) {
	type FakeInputVars struct {
		Zone               string
		Tags               string
		Project            string
		GCPCredentialsJSON string
		ExternalIP         string
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
				Zone:               "",
				Tags:               "",
				Project:            "",
				GCPCredentialsJSON: "",
				ExternalIP:         "",
			},
			args:    "",
			want:    "",
			wantErr: false,
		},
		{name: "Failure",
			fakeInputVars: FakeInputVars{},
			args:          "{{ .FakeKey }} \n",
			want:          "",
			wantErr:       true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v := &gcp.InputVars{
				Zone:               test.fakeInputVars.Zone,
				Tags:               test.fakeInputVars.Tags,
				Project:            test.fakeInputVars.Project,
				GCPCredentialsJSON: test.fakeInputVars.GCPCredentialsJSON,
				ExternalIP:         test.fakeInputVars.ExternalIP,
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
		Network gcp.MetadataStringValue
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
			Network: gcp.MetadataStringValue{
				Value: "fakeNetwork",
			},
		},
		want:    "fakeNetwork",
		fakeKey: "Network",
	},
		{
			name: "Failure",
			fields: fields{
				Network: gcp.MetadataStringValue{
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
			metadata := &gcp.Metadata{
				Network: test.fields.Network,
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
		fakeInputVars gcp.InputVars
		name          string
		data          map[string]interface{}
		wantErr       bool
	}{
		{
			name: "Success",
			data: map[string]interface{}{
				"Region":             "aRegion",
				"Zone":               "aZone",
				"Tags":               "someTags",
				"Project":            "aProject",
				"GCPCredentialsJSON": "aCredentialsJSONfile",
				"ExternalIP":         "anExternalIP",
				"Deployment":         "aDeployment",
				"ConfigBucket":       "aConfigBucket",
				"DBName":             "aDBName",
				"DBUsername":         "aDBUsername",
				"DBPassword":         "aDBPassword",
				"DBTier":             "aDBTier",
				"AllowIPs":           "aAllowIPs",
				"DNSManagedZoneName": "aDNSManagedZoneName",
				"DNSRecordSetPrefix": "aDNSRecordSetPrefix",
			},
			fakeInputVars: gcp.InputVars{},
		},
		{
			name: "Failure",
			data: map[string]interface{}{
				"Region":             "aRegion",
				"Zone":               "aZone",
				"Tags":               "someTags",
				"Project":            "aProject",
				"GCPCredentialsJSON": "aCredentialsJSONfile",
				"ExternalIP":         "anExternalIP",
				"Deployment":         "aDeployment",
			},
			wantErr:       true,
			fakeInputVars: gcp.InputVars{},
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
			data:          `{"director_public_ip":{"sensitive":false,"type": "string","value": "fakeIP"}}`,
			keyToSet:      "DirectorPublicIP",
			expectedValue: "fakeIP",
		},
		{
			name:          "Failure",
			keyToSet:      "DirectorPublicIP",
			expectedValue: "",
			wantErr:       true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata := &gcp.Metadata{}
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
