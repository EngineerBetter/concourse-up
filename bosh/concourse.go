package bosh

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/EngineerBetter/concourse-up/db"
)

const concourseManifestFilename = "concourse.yml"
const credsFilename = "concourse-creds.yml"
const concourseDeploymentName = "concourse"
const concourseVersionsFilename = "versions.json"
const concourseGrafanaFilename = "grafana_dashboard.yml"
const concourseCompatibilityFilename = "cup_compatibility.yml"
const postgresCACertFilename = "ca.pem"

func (client *Client) uploadConcourseStemcell() error {
	var ops []struct {
		Path  string
		Value json.RawMessage
	}
	err := json.Unmarshal(awsConcourseVersions, &ops)
	if err != nil {
		return err
	}
	var version string
	for _, op := range ops {
		if op.Path != "/stemcells/alias=trusty/version" {
			continue
		}
		err := json.Unmarshal(op.Value, &version)
		if err != nil {
			return err
		}
	}
	if version == "" {
		return errors.New("did not find stemcell version in versions.json")
	}
	return client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		false,
		"upload-stemcell",
		fmt.Sprintf("https://s3.amazonaws.com/bosh-aws-light-stemcells/light-bosh-stemcell-%s-aws-xen-hvm-ubuntu-trusty-go_agent.tgz", version),
	)
}

func (client *Client) deployConcourse(creds []byte, detach bool) (newCreds []byte, err error) {

	concourseManifestPath, err := client.director.SaveFileToWorkingDir(concourseManifestFilename, awsConcourseManifest)
	if err != nil {
		return
	}

	concourseVersionsPath, err := client.director.SaveFileToWorkingDir(concourseVersionsFilename, awsConcourseVersions)
	if err != nil {
		return
	}

	concourseCompatibilityPath, err := client.director.SaveFileToWorkingDir(concourseCompatibilityFilename, awsConcourseCompatibility)
	if err != nil {
		return
	}

	concourseGrafanaPath, err := client.director.SaveFileToWorkingDir(concourseGrafanaFilename, awsConcourseGrafana)
	if err != nil {
		return
	}

	postgresCACertPath, err := client.director.SaveFileToWorkingDir(postgresCACertFilename, []byte(db.RDSRootCert))
	if err != nil {
		return
	}

	credsPath, err := client.director.SaveFileToWorkingDir(credsFilename, creds)
	if err != nil {
		return
	}

	err = client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		detach,
		"--deployment",
		concourseDeploymentName,
		"deploy",
		concourseManifestPath,
		"--vars-store",
		credsPath,
		"--ops-file",
		concourseVersionsPath,
		"--ops-file",
		concourseCompatibilityPath,
		"--vars-file",
		concourseGrafanaPath,
		"--var",
		"deployment_name="+concourseDeploymentName,
		"--var",
		"domain="+client.config.Domain,
		"--var",
		"project="+client.config.Project,
		"--var",
		"web_network_name=public",
		"--var",
		"worker_network_name=private",
		"--var-file",
		"postgres_ca_cert="+postgresCACertPath,
		"--var",
		"postgres_host="+client.metadata.BoshDBAddress.Value,
		"--var",
		"postgres_port="+client.metadata.BoshDBPort.Value,
		"--var",
		"postgres_role="+client.config.RDSUsername,
		"--var",
		"postgres_password="+client.config.RDSPassword,
		"--var",
		"postgres_host="+client.metadata.BoshDBAddress.Value,
		"--var",
		"web_vm_type=concourse-web-"+client.config.ConcourseWebSize,
		"--var",
		"worker_vm_type=concourse-"+client.config.ConcourseWorkerSize,
		"--var",
		"atc_eip="+client.metadata.ATCPublicIP.Value,
	)
	newCreds, err1 := ioutil.ReadFile(credsPath)
	if err == nil {
		err = err1
	}
	return
}

//go:generate go-bindata -pkg $GOPACKAGE  assets/...
var awsConcourseManifest = MustAsset("assets/manifest.yml")
var awsConcourseGrafana = MustAsset("assets/grafana_dashboard.yml")
var awsConcourseVersions = MustAsset("assets/ops/versions.json")
var awsConcourseCompatibility = MustAsset("assets/ops/cup_compatibility.yml")
