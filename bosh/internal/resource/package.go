package resource

import (
	"encoding/json"

	"github.com/EngineerBetter/concourse-up/bosh/internal/resource/internal/file"
)

//go:generate go-bindata -o internal/file/file.go -ignore (\.go$)|(\.git) -nometadata -pkg file -prefix=../../../../concourse-up-ops . ../../../../concourse-up-ops/...

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
	DirectorManifest  = mustAssetString("director/director-manifest.yml")
	AWSCPIOps         = mustAssetString("director/aws-cpi.yml")
	ExternalIPOps     = mustAssetString("director/external-ip.yml")
	DirectorCustomOps = mustAssetString("director/custom-ops.yml")
)

// NOTE(px) remove this in a later version of github.com/mattn/go-bindata
func mustAssetString(name string) string {
	return string(file.MustAsset(name))
}

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
