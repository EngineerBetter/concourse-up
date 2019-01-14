package iaas

import (
	"testing"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

// GCP authorisation requires that the environment variable $GOOGLE_APPLICATION_CREDENTIALS be set

var ctx = context.Background()

// IAAS represents actions taken against GCP
type IAAS interface {
	BucketExists(string) (bool, error)
	CreateBucket(string) error
	DeleteBucket(string) error
	DeleteFile(string, string) error
	EnsureFileExists(string, string, string) ([]byte, bool, error)
	HasFile(string string) (bool, error)
	LoadFile(string, string) ([]byte, error)
	WriteFile(string, string) error
}

func TestGCPProvider_IAAS(t *testing.T) {
	type fields struct {
		ctx     context.Context
		storage *storage.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "returns the proper IAAS name",
			fields: fields{},
			want:   "GCP",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GCPProvider{
				ctx:     tt.fields.ctx,
				storage: tt.fields.storage,
			}
			if got := g.IAAS(); got != tt.want {
				t.Errorf("GCPProvider.IAAS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGCPProvider_Region(t *testing.T) {
	type fields struct {
		ctx     context.Context
		storage *storage.Client
		region  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name: "returns defined region",
		fields: fields{
			region: "aRegion",
		},
		want: "aRegion",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GCPProvider{
				ctx:     tt.fields.ctx,
				storage: tt.fields.storage,
				region:  tt.fields.region,
			}
			if got := g.Region(); got != tt.want {
				t.Errorf("GCPProvider.Region() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGCPProvider_DBType(t *testing.T) {
	type fields struct {
		ctx     context.Context
		storage GCPStorageClient
		region  string
		attrs   map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		size   string
		want   string
	}{
		{
			name:   "Success: correctly maps 'medium' size",
			fields: fields{},
			size:   "medium",
			want:   "db-custom-2-4096",
		},
		{
			name:   "Success: correctly maps 'small' size",
			fields: fields{},
			size:   "small",
			want:   "db-g1-small",
		},
		{
			name:   "Success: correctly maps 'large' size",
			fields: fields{},
			size:   "large",
			want:   "db-custom-2-8192",
		},
		{
			name:   "Success: correctly maps 'xlarge' size",
			fields: fields{},
			size:   "xlarge",
			want:   "db-custom-4-16384",
		},
		{
			name:   "Success: correctly maps '2xlarge' size",
			fields: fields{},
			size:   "2xlarge",
			want:   "db-custom-8-32768",
		},
		{
			name:   "Success: correctly maps '4xlarge' size",
			fields: fields{},
			size:   "4xlarge",
			want:   "db-custom-16-65536",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GCPProvider{
				ctx:     tt.fields.ctx,
				storage: tt.fields.storage,
				region:  tt.fields.region,
				attrs:   tt.fields.attrs,
			}
			if got := g.DBType(tt.size); got != tt.want {
				t.Errorf("GCPProvider.DBType() = %v, want %v", got, tt.want)
			}
		})
	}
}
