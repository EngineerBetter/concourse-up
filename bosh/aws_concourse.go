package bosh

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/EngineerBetter/concourse-up/db"
)

func (client *AWSClient) deployConcourse(creds []byte, detach bool) ([]byte, error) {

	err := saveFilesToWorkingDir(client.workingdir, client.provider, creds)
	if err != nil {
		return nil, fmt.Errorf("failed saving files to working directory in deployConcourse: [%v]", err)
	}

	boshDBAddress, err := client.outputs.Get("BoshDBAddress")
	if err != nil {
		return nil, err
	}
	boshDBPort, err := client.outputs.Get("BoshDBPort")
	if err != nil {
		return nil, err
	}
	atcPublicIP, err := client.outputs.Get("ATCPublicIP")
	if err != nil {
		return nil, err
	}

	vmap := map[string]interface{}{
		"deployment_name":          concourseDeploymentName,
		"domain":                   client.config.Domain,
		"project":                  client.config.Project,
		"web_network_name":         "public",
		"worker_network_name":      "private",
		"postgres_host":            boshDBAddress,
		"postgres_port":            boshDBPort,
		"postgres_role":            client.config.RDSUsername,
		"postgres_password":        client.config.RDSPassword,
		"postgres_ca_cert":         db.RDSRootCert,
		"web_vm_type":              "concourse-web-" + client.config.ConcourseWebSize,
		"worker_vm_type":           "concourse-" + client.config.ConcourseWorkerSize,
		"worker_count":             client.config.ConcourseWorkerCount,
		"atc_eip":                  atcPublicIP,
		"external_tls.certificate": client.config.ConcourseCert,
		"external_tls.private_key": client.config.ConcourseKey,
		"atc_encryption_key":       client.config.EncryptionKey,
	}

	flagFiles := []string{
		client.workingdir.PathInWorkingDir(concourseManifestFilename),
		"--vars-store",
		client.workingdir.PathInWorkingDir(credsFilename),
		"--ops-file",
		client.workingdir.PathInWorkingDir(concourseVersionsFilename),
		"--ops-file",
		client.workingdir.PathInWorkingDir(concourseSHAsFilename),
		"--ops-file",
		client.workingdir.PathInWorkingDir(concourseCompatibilityFilename),
		"--vars-file",
		client.workingdir.PathInWorkingDir(concourseGrafanaFilename),
	}

	if client.config.ConcoursePassword != "" {
		vmap["atc_password"] = client.config.ConcoursePassword
	}

	if client.config.GithubAuthIsSet {
		vmap["github_client_id"] = client.config.GithubClientID
		vmap["github_client_secret"] = client.config.GithubClientSecret
		flagFiles = append(flagFiles, "--ops-file", client.workingdir.PathInWorkingDir(concourseGitHubAuthFilename))
	}

	t, err1 := client.buildTagsYaml(vmap["project"], "concourse")
	if err1 != nil {
		return nil, err
	}
	vmap["tags"] = t
	flagFiles = append(flagFiles, "--ops-file", client.workingdir.PathInWorkingDir(extraTagsFilename))

	vs := vars(vmap)

	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve director IP: [%v]", err)
	}

	err = client.boshCLI.RunAuthenticatedCommand(
		"deploy",
		directorPublicIP,
		client.config.DirectorPassword,
		client.config.DirectorCACert,
		detach,
		os.Stdout,
		append(flagFiles, vs...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to run bosh deploy with commands %+v: [%v]", flagFiles, err)
	}

	return ioutil.ReadFile(client.workingdir.PathInWorkingDir(credsFilename))
}

func (client *AWSClient) buildTagsYaml(project interface{}, component string) (string, error) {
	var b strings.Builder

	for _, e := range client.config.Tags {
		kv := strings.Join(strings.Split(e, "="), ": ")
		_, err := fmt.Fprintf(&b, "%s,", kv)
		if err != nil {
			return "", err
		}
	}
	cProjectTag := fmt.Sprintf("concourse-up-project: %v,", project)
	b.WriteString(cProjectTag)
	cComponentTag := fmt.Sprintf("concourse-up-component: %s", component)
	b.WriteString(cComponentTag)
	return fmt.Sprintf("{%s}", b.String()), nil
}
