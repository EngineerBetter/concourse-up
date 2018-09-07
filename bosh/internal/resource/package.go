package resource

import (
	"encoding/json"
	"runtime"

	"github.com/EngineerBetter/concourse-up/bosh/internal/bincache"
	"github.com/EngineerBetter/concourse-up/bosh/internal/resource/internal/file"
)

//go:generate go-bindata -o internal/file/file.go -ignore (\.go$)|(\.git)|(bosh/assets) -nometadata -pkg file -prefix=../../../../concourse-up-ops . ../../../../concourse-up-ops/...

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
	DirectorManifest = mustAssetString("director/manifest.yml")
	// AWSCPIOps statically defines aws-cpi.yml contents
	AWSCPIOps = mustAssetString("director/aws/cpi.yml")
	// ExternalIPOps statically defines external-ip.yml contents
	ExternalIPOps = mustAssetString("director/external-ip.yml")
	// DirectorCustomOps statically defines custom-ops.yml contents
	DirectorCustomOps = mustAssetString("director/custom-ops.yml")
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
	p := file.MustAsset("director-versions.json")
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
	Windows string `json:"windows"`
	Mac     string `json:"mac"`
	Linux   string `json:"linux"`
}

func (p binaryPaths) path() string {
	switch runtime.GOOS {
	case "windows":
		return p.Windows
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
