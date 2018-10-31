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
