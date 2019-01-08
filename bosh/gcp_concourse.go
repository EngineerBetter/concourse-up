package bosh

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func (client *GCPClient) deployConcourse(creds []byte, detach bool) (newCreds []byte, err error) {

	concourseManifestPath, err := client.director.SaveFileToWorkingDir(concourseManifestFilename, awsConcourseManifest)
	if err != nil {
		return []byte{}, err
	}

	concourseVersionsPath, err := client.director.SaveFileToWorkingDir(concourseVersionsFilename, awsConcourseVersions)
	if err != nil {
		return []byte{}, err
	}

	concourseSHAsPath, err := client.director.SaveFileToWorkingDir(concourseSHAsFilename, awsConcourseSHAs)
	if err != nil {
		return []byte{}, err
	}

	concourseCompatibilityPath, err := client.director.SaveFileToWorkingDir(concourseCompatibilityFilename, awsConcourseCompatibility)
	if err != nil {
		return []byte{}, err
	}

	concourseGrafanaPath, err := client.director.SaveFileToWorkingDir(concourseGrafanaFilename, awsConcourseGrafana)
	if err != nil {
		return []byte{}, err
	}

	concourseGitHubAuthPath, err := client.director.SaveFileToWorkingDir(concourseGitHubAuthFilename, awsConcourseGitHubAuth)
	if err != nil {
		return []byte{}, err
	}

	uaaCertPath, err := client.director.SaveFileToWorkingDir(uaaCertFilename, uaaCert)
	if err != nil {
		return []byte{}, err
	}

	credsPath, err := client.director.SaveFileToWorkingDir(credsFilename, creds)
	if err != nil {
		return []byte{}, err
	}

	extraTagsPath, err := client.director.SaveFileToWorkingDir(extraTagsFilename, extraTags)
	if err != nil {
		return []byte{}, err
	}

	boshDBAddress, err := client.metadata.Get("BoshDBAddress")
	if err != nil {
		return []byte{}, err
	}
	atcPublicIP, err := client.metadata.Get("ATCPublicIP")
	if err != nil {
		return []byte{}, err
	}

	networkName, err := client.metadata.Get("Network")
	if err != nil {
		return []byte{}, err
	}

	SQLServerCert, err := client.metadata.Get("SQLServerCert")
	if err != nil {
		return []byte{}, err
	}

	vmap := map[string]interface{}{
		"deployment_name":          concourseDeploymentName,
		"domain":                   client.config.Domain,
		"project":                  client.config.Project,
		"web_network_name":         "public",
		"worker_network_name":      "private",
		"postgres_host":            boshDBAddress,
		"postgres_role":            client.config.RDSUsername,
		"postgres_port":            "5432",
		"postgres_password":        client.config.RDSPassword,
		"postgres_ca_cert":         SQLServerCert,
		"web_vm_type":              "concourse-web-" + client.config.ConcourseWebSize,
		"worker_vm_type":           "concourse-" + client.config.ConcourseWorkerSize,
		"worker_count":             client.config.ConcourseWorkerCount,
		"atc_eip":                  atcPublicIP,
		"external_tls.certificate": client.config.ConcourseCert,
		"external_tls.private_key": client.config.ConcourseKey,
		"atc_encryption_key":       client.config.EncryptionKey,
		"network_name":             networkName,
	}

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
		"--ops-file",
		uaaCertPath,
		"--vars-file",
		concourseGrafanaPath,
	}

	if client.config.ConcoursePassword != "" {
		vmap["atc_password"] = client.config.ConcoursePassword
	}

	if client.config.GithubAuthIsSet {
		vmap["github_client_id"] = client.config.GithubClientID
		vmap["github_client_secret"] = client.config.GithubClientSecret
		flagFiles = append(flagFiles, "--ops-file", concourseGitHubAuthPath)
	}

	t, err1 := client.buildTagsYaml(vmap["project"], "concourse")
	if err1 != nil {
		return []byte{}, err
	}
	vmap["tags"] = t
	flagFiles = append(flagFiles, "--ops-file", extraTagsPath)

	vs := vars(vmap)

	err = client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		detach,
		append(flagFiles, vs...)...,
	)
	newCreds, err1 = ioutil.ReadFile(credsPath)
	if err == nil {
		err = err1
	}
	return []byte{}, err
}

func (client *GCPClient) buildTagsYaml(project interface{}, component string) (string, error) {
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
