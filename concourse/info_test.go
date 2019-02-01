package concourse

import (
	"strings"
	"testing"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/config"
)

func TestInfo_String(t *testing.T) {
	type fields struct {
		Terraform   TerraformInfo
		Config      config.Config
		Instances   []bosh.Instance
		CertExpiry  string
		GatewayUser string
	}
	defaultFields := fields{
		Terraform: TerraformInfo{
			DirectorPublicIP: "4.3.2.1",
			NatGatewayIP:     "1.2.3.4",
		},
		Config:      config.Config{},
		Instances:   []bosh.Instance{},
		CertExpiry:  "2019-02-01",
		GatewayUser: "gateway user",
	}
	tests := []struct {
		name   string
		fields fields
		init   func(fields) fields
		want   string
	}{
		{
			name:   "basic TestInfo templating",
			fields: defaultFields,
			init: func(f fields) fields {
				return f
			},
			want: "Outbound Public IP: 1.2.3.4",
		},
		{
			name:   "iaas templating",
			fields: defaultFields,
			init: func(f fields) fields {
				f.Config.IAAS = "aCloudProvider"
				return f
			},
			want: "IAAS:      aCloudProvider",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields = tt.init(tt.fields)
			info := &Info{
				Terraform:   tt.fields.Terraform,
				Config:      tt.fields.Config,
				Instances:   tt.fields.Instances,
				CertExpiry:  tt.fields.CertExpiry,
				GatewayUser: tt.fields.GatewayUser,
			}
			if got := info.String(); !strings.Contains(got, tt.want) {
				t.Errorf("Info.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
