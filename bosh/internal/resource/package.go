package resource

import (
	"encoding/json"

	"github.com/EngineerBetter/concourse-up/bosh/internal/resource/internal/file"
)

//go:generate go-bindata -o internal/file/files.go -ignore \.git -nometadata -pkg file -prefix=../../../../concourse-up-ops ../../../../concourse-up-ops/...

type Resource struct {
	URL     string `json:"url"`
	Version string `json:"version"`
	SHA1    string `json:"sha1"`
}

var resources map[string]Resource

type ResourceID struct {
	name string
}

var (
	AWSCPI      = ResourceID{"cpi"}
	AWSStemcell = ResourceID{"stemcell"}
	BOSHRelease = ResourceID{"bosh"}
	BPMRelease  = ResourceID{"bpm"}
)

var (
	DirectorManifest  = file.MustAssetString("director/manifest.yml")
	AWSCPIOps         = file.MustAssetString("director/aws/cpi.yml")
	ExternalIPOps     = file.MustAssetString("director/external-ip.yml")
	DirectorCustomOps = file.MustAssetString("director/custom-ops.yml")
)

func Get(id ResourceID) Resource {
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
}
