package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ghodss/yaml"
)

func main() {
	f, err := os.Create(os.Getenv("TARGET_FILE"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	x := make(map[string]interface{})

	x["fly"], err = getFlyURL()
	if err != nil {
		log.Fatal(err)
	}
	x["bosh-cli"] = getBOSHCLIURL()
	x["terraform"] = getTerraformURL()
	x["bosh"], err = getBOSH()
	if err != nil {
		log.Fatal(err)
	}
	x["cpi"], err = getCPI()
	if err != nil {
		log.Fatal(err)
	}
	x["bpm"], err = getBPM()
	if err != nil {
		log.Fatal(err)
	}
	x["stemcell"], err = getStemcell()
	if err != nil {
		log.Fatal(err)
	}

	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	err = e.Encode(x)
	if err != nil {
		log.Fatal(err)
	}
}

type multiPlatform struct {
	Windows string `json:"windows"`
	Mac     string `json:"mac"`
	Linux   string `json:"linux"`
}

type op struct {
	Path  string
	Value json.RawMessage
}

type opsFile []op

func (os opsFile) unmarshalPath(p string, v interface{}) error {
	var m json.RawMessage
	for _, o := range os {
		if o.Path != p {
			continue
		}
		m = o.Value
	}
	if m == nil {
		return errors.New("could not find op")
	}
	return yaml.Unmarshal(m, v)
}

func readOpsFile(path string) (opsFile, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var x opsFile
	err = yaml.Unmarshal(f, &x)
	return x, err
}

func getFlyURL() (multiPlatform, error) {
	ops, err := readOpsFile("concourse-up-manifest/ops/versions.json")
	if err != nil {
		return multiPlatform{}, err
	}
	var x struct {
		Version string `json:"version"`
	}
	err = ops.unmarshalPath("/releases/name=concourse", &x)
	if err != nil {
		return multiPlatform{}, err
	}
	return multiPlatform{
		Windows: fmt.Sprintf("https://github.com/concourse/concourse/releases/download/v%s/fly_windows_amd64", x.Version),
		Mac:     fmt.Sprintf("https://github.com/concourse/concourse/releases/download/v%s/fly_darwin_amd64", x.Version),
		Linux:   fmt.Sprintf("https://github.com/concourse/concourse/releases/download/v%s/fly_linux_amd64", x.Version),
	}, nil
}

func getTerraformURL() multiPlatform {
	return multiPlatform{
		Windows: "https://releases.hashicorp.com/terraform/0.11.7/terraform_0.11.7_windows_amd64.zip",
		Mac:     "https://releases.hashicorp.com/terraform/0.11.7/terraform_0.11.7_darwin_amd64.zip",
		Linux:   "https://releases.hashicorp.com/terraform/0.11.7/terraform_0.11.7_linux_amd64.zip",
	}
}

func getBOSHCLIURL() multiPlatform {
	return multiPlatform{
		Windows: "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.0.1-windows-amd64.exe",
		Mac:     "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.0.1-darwin-amd64",
		Linux:   "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.0.1-linux-amd64",
	}
}

type resource struct {
	URL     string `json:"url"`
	SHA1    string `json:"sha1"`
	Version string `json:"version"`
}

func getResource(path string) (resource, error) {
	ops, err := readOpsFile("bosh-deployment/aws/cpi.yml")
	if err != nil {
		return resource{}, err
	}
	var x resource
	err = ops.unmarshalPath(path, &x)
	return x, err
}

func getStemcell() (resource, error) {
	return getResource("/resource_pools/name=vms/stemcell?")
}

func getCPI() (resource, error) {
	return getResource("/releases/-")
}

func getRelease(file, name string) (resource, error) {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return resource{}, err
	}
	var manifest struct {
		Releases []struct {
			Name string `json:"name"`
			resource
		} `json:"releases"`
	}
	err = yaml.Unmarshal(f, &manifest)
	if err != nil {
		return resource{}, err
	}
	for _, r := range manifest.Releases {
		if r.Name != name {
			continue
		}
		return r.resource, nil
	}
	return resource{}, errors.New("release not found")
}

func getBOSH() (resource, error) {
	return getRelease("bosh-deployment/bosh.yml", "bosh")
}

func getBPM() (resource, error) {
	return getRelease("bosh-deployment/bosh.yml", "bpm")
}
