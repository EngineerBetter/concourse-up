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
const concourseSHAsFilename = "shas.json"
const concourseGrafanaFilename = "grafana_dashboard.yml"
const concourseCompatibilityFilename = "cup_compatibility.yml"
const concourseGitHubAuthFilename = "github-auth.yml"

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
		if op.Path != "/stemcells/alias=xenial/version" {
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
		fmt.Sprintf("https://s3.amazonaws.com/bosh-aws-light-stemcells/light-bosh-stemcell-%s-aws-xen-hvm-ubuntu-xenial-go_agent.tgz", version),
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

	concourseSHAsPath, err := client.director.SaveFileToWorkingDir(concourseSHAsFilename, awsConcourseSHAs)
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

	concourseGitHubAuthPath, err := client.director.SaveFileToWorkingDir(concourseGitHubAuthFilename, awsConcourseGitHubAuth)
	if err != nil {
		return
	}

	credsPath, err := client.director.SaveFileToWorkingDir(credsFilename, creds)
	if err != nil {
		return
	}

	vmap := map[string]interface{}{
		"deployment_name":          concourseDeploymentName,
		"domain":                   client.config.Domain,
		"project":                  client.config.Project,
		"web_network_name":         "public",
		"worker_network_name":      "private",
		"postgres_host":            client.metadata.BoshDBAddress.Value,
		"postgres_port":            client.metadata.BoshDBPort.Value,
		"postgres_role":            client.config.RDSUsername,
		"postgres_password":        client.config.RDSPassword,
		"postgres_ca_cert":         db.RDSRootCert,
		"web_vm_type":              "concourse-web-" + client.config.ConcourseWebSize,
		"worker_vm_type":           "concourse-" + client.config.ConcourseWorkerSize,
		"worker_count":             client.config.ConcourseWorkerCount,
		"atc_eip":                  client.metadata.ATCPublicIP.Value,
		"external_tls.certificate": client.config.ConcourseCert,
		"external_tls.private_key": client.config.ConcourseKey,
		"atc_encryption_key":       client.config.EncryptionKey,
	}

	if client.config.ConcoursePassword != "" {
		vmap["atc_password"] = client.config.ConcoursePassword
	}

	if client.config.GithubAuthIsSet {
		vmap["github_client_id"] = client.config.GithubClientID
		vmap["github_client_secret"] = client.config.GithubClientSecret
	}

	vs := vars(vmap)

	flagFiles := []string{
		"--deployment",
		concourseDeploymentName,
		"deploy",
		concourseManifestPath,
		"--vars-store",
		credsPath,
		"--ops-file",
		concourseVersionsPath,
		"--ops-file",
		concourseSHAsPath,
		"--ops-file",
		concourseCompatibilityPath,
		"--vars-file",
		concourseGrafanaPath,
	}
	if client.config.GithubAuthIsSet {
		flagFiles = append(flagFiles, "--ops-file", concourseGitHubAuthPath)
	}

	err = client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		detach,
		append(flagFiles, vs...)...,
	)
	newCreds, err1 := ioutil.ReadFile(credsPath)
	if err == nil {
		err = err1
	}
	return
}

func vars(vars map[string]interface{}) []string {
	var x []string
	for k, v := range vars {
		switch v.(type) {
		case string:
			x = append(x, "--var", fmt.Sprintf("%s=%q", k, v))
		case int:
			x = append(x, "--var", fmt.Sprintf("%s=%d", k, v))
		default:
			panic("unsupported type")
		}
	}
	return x
}

//go:generate go-bindata -pkg $GOPACKAGE  assets/... ../../concourse-up-ops/...
var awsConcourseGrafana = MustAsset("assets/grafana_dashboard.yml")
var awsConcourseCompatibility = MustAsset("assets/ops/cup_compatibility.yml")
var awsConcourseGitHubAuth = MustAsset("assets/ops/github-auth.yml")
var awsConcourseManifest = MustAsset("../../concourse-up-ops/manifest.yml")
var awsConcourseVersions = MustAsset("../../concourse-up-ops/ops/versions.json")
var awsConcourseSHAs = MustAsset("../../concourse-up-ops/ops/shas.json")
