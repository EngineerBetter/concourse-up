package resource

import (
	"encoding/json"
	"runtime"

	"github.com/EngineerBetter/concourse-up/resource/internal/file"
	"github.com/EngineerBetter/concourse-up/util/bincache"
)

//go:generate go-bindata -o internal/file/file.go -ignore (\.go$)|(\.git)|(bosh/assets) -nometadata -pkg file ../../concourse-up-ops/... assets/...

// Resource safely exposes the json parameters of a resource
type Resource struct {
	URL     string `json:"url"`
	Version string `json:"version"`
	SHA1    string `json:"sha1"`
}

var resources map[string]Resource

// ID defines the name of a resource in a safer way
type ID struct {
	name string
}

var (
	// AWSCPI statically defines cpi string
	AWSCPI = ID{"cpi"}
	// AWSStemcell statically defines stemcell string
	AWSStemcell = ID{"stemcell"}
	// BOSHRelease statically defines bosh string
	BOSHRelease = ID{"bosh"}
	// BPMRelease statically defines bpm string
	BPMRelease = ID{"bpm"}
)

var (
	// DirectorManifest statically defines director-manifest.yml contents
	DirectorManifest = mustAssetString("assets/manifest.yml")

	// AWSDirectorCloudConfig statically defines aws cloud-config.yml
	AWSDirectorCloudConfig = mustAssetString("assets/aws/cloud-config.yml")
	// AWSCPIOps statically defines aws-cpi.yml contents
	AWSCPIOps = mustAssetString("assets/aws/cpi.yml")
	//GCPJumpboxUserOps statically defines gcp jumpbox-user.yml
	GCPJumpboxUserOps = mustAssetString("assets/gcp/jumpbox-user.yml")
	// GCPDirectorCloudConfig statically defines gcp cloud-config.yml
	GCPDirectorCloudConfig = mustAssetString("assets/gcp/cloud-config.yml")
	// GCPCPIOps statically defines gcp-cpi.yml contents
	GCPCPIOps = mustAssetString("assets/gcp/cpi.yml")
	// GCPExternalIPOps statically defines external-ip.yml contents
	GCPExternalIPOps = mustAssetString("assets/gcp/external-ip.yml")
	// GCPDirectorCustomOps statically defines custom-ops.yml contents
	GCPDirectorCustomOps = mustAssetString("assets/gcp/custom-ops.yml")

	// AWSTerraformConfig holds the terraform conf for AWS
	AWSTerraformConfig = mustAssetString("assets/aws/infrastructure.tf")

	// GCPTerraformConfig holds the terraform conf for GCP
	GCPTerraformConfig = mustAssetString("assets/gcp/infrastructure.tf")

	// ExternalIPOps statically defines external-ip.yml contents
	ExternalIPOps = mustAssetString("assets/external-ip.yml")
	// AWSDirectorCustomOps statically defines custom-ops.yml contents
	AWSDirectorCustomOps = mustAssetString("assets/aws/custom-ops.yml")

	// AWSReleaseVersions carries all versions of releases
	AWSReleaseVersions = mustAssetString("../../concourse-up-ops/ops/versions-aws.json")

	// GCPReleaseVersions carries all versions of releases
	GCPReleaseVersions = mustAssetString("../../concourse-up-ops/ops/versions-gcp.json")

	// AddNewCa carries the ops file that adds a new CA required for cert rotation
	AddNewCa = mustAssetString("assets/maintenance/add-new-ca.yml")

	// RemoveOldCa carries the ops file that removes the old CA required for cert rotation
	RemoveOldCa = mustAssetString("assets/maintenance/remove-old-ca.yml")

	// CleanupCerts moves renewed values of certs to old keys in director vars store
	CleanupCerts = mustAssetString("assets/maintenance/cleanup-certs.yml")
)

// NOTE(px) remove this in a later version of github.com/mattn/go-bindata
func mustAssetString(name string) string {
	return string(file.MustAsset(name))
}

// Get returns an Resource in a safe way
func Get(id ID) Resource {
	r, ok := resources[id.name]
	if !ok {
		panic("resource " + id.name + " not found")
	}
	return r
}

func init() {
	p := file.MustAsset("../../concourse-up-ops/director-versions.json")
	err := json.Unmarshal(p, &resources)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(p, &binaries)
	if err != nil {
		panic(err)
	}
}

var binaries map[string]binaryPaths

type binaryPaths struct {
	Mac   string `json:"mac"`
	Linux string `json:"linux"`
}

func (p binaryPaths) path() string {
	switch runtime.GOOS {
	case "darwin":
		return p.Mac
	case "linux":
		return p.Linux
	default:
		panic("OS not supported")
	}
}

// BOSHCLIPath returns the path of the downloaded bosh-cli
func BOSHCLIPath() (string, error) {
	p := binaries["bosh-cli"].path()
	return bincache.Download(p)
}

// TerraformCLIPath returns the path of the downloaded terraform-cli
func TerraformCLIPath() (string, error) {
	p := binaries["terraform"].path()
	return bincache.Download(p)
}
