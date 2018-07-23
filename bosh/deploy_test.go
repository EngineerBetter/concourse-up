package bosh

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/EngineerBetter/concourse-up/terraform"

	"io/ioutil"

	"github.com/EngineerBetter/concourse-up/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deploy", func() {
	var actions []string
	var tempDir string
	var createEnvOutput string

	directorClient := &FakeDirectorClient{
		FakeRunCommand: func(stdout, stderr io.Writer, args ...string) error {
			actions = append(actions, fmt.Sprintf("Running bosh command: %s", strings.Join(args, " ")))

			err := ioutil.WriteFile(filepath.Join(tempDir, "director-state.json"), []byte("{ some state }"), 0700)
			Expect(err).ToNot(HaveOccurred())

			if strings.Contains(strings.Join(args, " "), "create-env") {
				_, err := stdout.Write([]byte(createEnvOutput))
				Expect(err).ToNot(HaveOccurred())
				return nil
			}
			return nil
		},
		FakeRunAuthenticatedCommand: func(stdout, stderr io.Writer, detach bool, args ...string) error {
			actions = append(actions, fmt.Sprintf("Running authenticated bosh command: %s (detach: %t)", strings.Join(args, " "), detach))
			return nil
		},
		FakeSaveFileToWorkingDir: func(filename string, contents []byte) (string, error) {
			actions = append(actions, fmt.Sprintf("Saving file to working dir: %s", filename))
			err := ioutil.WriteFile(filepath.Join(tempDir, filename), contents, 0700)
			Expect(err).ToNot(HaveOccurred())
			return filepath.Join(tempDir, filename), nil
		},
		FakePathInWorkingDir: func(filename string) string {
			return filepath.Join(tempDir, filename)
		},
		FakeCleanup: func() error {
			actions = append(actions, "Cleaning up")
			return nil
		},
	}

	var client IClient

	BeforeEach(func() {
		actions = []string{}
		createEnvOutput = "Finished deploying"

		var err error
		tempDir, err = ioutil.TempDir("", "bosh_test")
		Expect(err).ToNot(HaveOccurred())

		terraformMetadata := &terraform.Metadata{
			ATCPublicIP:              terraform.MetadataStringValue{Value: "77.77.77.77"},
			ATCSecurityGroupID:       terraform.MetadataStringValue{Value: "sg-888"},
			BlobstoreBucket:          terraform.MetadataStringValue{Value: "blobs.aws.com"},
			BlobstoreSecretAccessKey: terraform.MetadataStringValue{Value: "abc123"},
			BlobstoreUserAccessKeyID: terraform.MetadataStringValue{Value: "abc123"},
			BoshDBAddress:            terraform.MetadataStringValue{Value: "rds.aws.com"},
			BoshDBPort:               terraform.MetadataStringValue{Value: "5432"},
			BoshSecretAccessKey:      terraform.MetadataStringValue{Value: "abc123"},
			BoshUserAccessKeyID:      terraform.MetadataStringValue{Value: "abc123"},
			DirectorKeyPair:          terraform.MetadataStringValue{Value: "-- KEY --"},
			DirectorPublicIP:         terraform.MetadataStringValue{Value: "99.99.99.99"},
			DirectorSecurityGroupID:  terraform.MetadataStringValue{Value: "sg-123"},
			PrivateSubnetID:          terraform.MetadataStringValue{Value: "sn-private-123"},
			PublicSubnetID:           terraform.MetadataStringValue{Value: "sn-public-123"},
			VMsSecurityGroupID:       terraform.MetadataStringValue{Value: "sg-456"},
		}

		exampleConfig := &config.Config{
			PublicKey: "example-public-key",
			PrivateKey: `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAsoHoo0qYchdIiXIOB4EEWo060NrgUqIIH+o8KLOPXfVBnffS
dCX1tpeOJd5qwou/YsEBvBuA7oX+qymT9Y+AOf0l9ck8zCzuHxHyYdoK31orTHax
jZMjLCYPj/Dffa50IQH27ntKSFrxW0PlWSkjb9W5rO7YXUhe41Ut19MP/0pQEZ/H
ZziQpI8Jgk2RVVAl4ffKJFRyfXd/iBvn+lBY+y2QYL1gVb53BPEn2F88Z9hUkwEJ
TbeIPxuw8tHDKIj4aJs56sRhkZtLQyoNiQlHMU8FVXnh0dTPoFIkPHzvgKV9ywO8
sLVTrbKl7MLKl8Y7WCAx0Gh6+YaCU/nHksM0kQIDAQABAoIBABKuq/VjGj9elnXk
HPnGE/mSLGStc6rSUH1em3s7B7cysvJgfIMxcdzxUaw+8fd4fshMIO1aB41vMq8h
Q94AbdAj4XQu4pEP5sATtcVt95NWsY9oIL8LdjPpq9lJwWo69uZ5eSmOd8DI29fM
bFV/i7jpqmwh9z0UFPI/+PNMoLD8HlNJslnBWDAUWvuE+h43cmx7k0pUCx5vP3Ew
moyNppYSpd5uskyxEZ0r8s3IZW43ipxXdN0oL9zuj0ra69fVGtDikEFdpgtDMpmi
hhzrE6yjxFhmzI2PaPbvYAp90pUVxXniXuZRaCGHo3nezP0KU8uoeGCnLEtwTgcL
GMeV1MECgYEA2ESqGrAthDyYsWcw6j+pqLnED8PrTwvTG3qZQ2+mTAOZL5KG+hjb
emPsWpPnuT+VFlaqqutt2PR9MaoFMDqt9ZegrgcOdJlLWuegJzEv2rpmELlgeGgF
pl0KrZ5fSk8CnZGYyZ2WGwO1gZY2j9cMrpYLRuz5vaan8d+Eerru8WkCgYEA001Q
O/tks3LzzcprrfHfpOitjzwLoIiVDjr34n4Ko1C8bZq5ANp9KFziBHqA31wuKXN2
FfQ4QQjD4v8ddvImThj/lHBsO/yO3vigo5e9VaIdIkl/YWmpS7yRI+oS2yquUCtj
1C7EXB+aWN7EZ/6YTQxFNmOBXaQ2LosIu8eSHOkCgYAgWJTAjR0hrBaCYha00nTD
oZUrbnghSHl4oKuPpIFQ2TDuJpI9kb4x3gQZwAlmcZYQ00GPcsrpKhgXd4BzKDOg
id8kaDXHRq44mHAhrH+lzT86vR8qoxRFP6E7OnayHIMdogsiDInI3JMnIJpkhRuG
eTaSkxr/PI/d4zpjSNY4EQKBgC9m0q8CEG8pRIRP+qQE9LTb9cOCJuGWgkm09NL8
j4pfnEXCReppGVaqr5Ftoed5mGl4G2+FX/FG9BrCPGvomqs+dGdqaP10BOEESZUp
fzHssjh04HyL5Yy1+qFh62T7SCt38GczLp20AT4ai1kBBk2SiRxQaj8FjZoXWpg1
hxOxAoGAQmevBK8NcUTnwfQ71sFulnfi5B3J0PdPzKb186vJnQoWjSdx5oceq+96
H6Xmkaua78D9NZSacQeHThBCeRWlykyDz0C20x5BnBl0PD86zbyxdqAFhCAU3T2n
X9M1dN0p4Xj/+GYmJTCPbrYm3Jb9BoaE49tJOc789M+VI7lPZ4s=
-----END RSA PRIVATE KEY-----`,
			Region:                 "eu-west-1",
			Deployment:             "concourse-up-happymeal",
			Project:                "happymeal",
			TFStatePath:            "example-path",
			DirectorUsername:       "admin",
			DirectorPassword:       "secret123",
			ConcourseDBName:        "concourse_atc",
			RDSUsername:            "admin",
			RDSPassword:            "s3cret",
			RDSDefaultDatabaseName: "default",
		}

		dbOpener := make(fakeOpener)
		db, mock, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		mock.ExpectExec("CREATE DATABASE concourse_atc").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("CREATE DATABASE uaa").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("CREATE DATABASE credhub").WillReturnResult(sqlmock.NewResult(0, 0))
		dbOpener[exampleConfig.RDSDefaultDatabaseName] = db
		db, mock, err = sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		mock.ExpectExec("TRUNCATE").WillReturnResult(sqlmock.NewResult(0, 0))
		dbOpener["credhub"] = db
		client = &Client{
			config:   exampleConfig,
			metadata: terraformMetadata,
			director: directorClient,
			db:       dbOpener,
			stdout:   new(bytes.Buffer),
			stderr:   new(bytes.Buffer),
		}
	})

	Context("When an initial director state exists", func() {
		It("Saves the director state", func() {
			stateFileBytes := []byte("{}")
			_, _, err := client.Deploy(stateFileBytes, nil, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(actions).To(ContainElement("Saving file to working dir: director-state.json"))
		})
	})

	Context("When an initial director state does not exist", func() {
		It("Does not save the director state", func() {
			_, _, err := client.Deploy(nil, nil, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(actions).ToNot(ContainElement("Saving file to working dir: director-state.json"))
		})
	})

	It("Saves the private key", func() {
		_, _, err := client.Deploy(nil, nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Saving file to working dir: director.pem"))
	})

	It("Saves the manifest", func() {
		_, _, err := client.Deploy(nil, nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Saving file to working dir: director.yml"))
	})

	It("Deploys the director", func() {
		_, _, err := client.Deploy(nil, nil, false)
		Expect(err).ToNot(HaveOccurred())
		expectedCommand := fmt.Sprintf("Running bosh command: create-env %s/director.yml --state %s/director-state.json --vars-store %s/director-creds.yml", tempDir, tempDir, tempDir)
		Expect(actions).To(ContainElement(expectedCommand))
	})

	It("Saves the cloud config", func() {
		_, _, err := client.Deploy(nil, nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Saving file to working dir: cloud-config.yml"))
	})

	It("Updates the cloud config", func() {
		_, _, err := client.Deploy(nil, nil, false)
		Expect(err).ToNot(HaveOccurred())
		expectedCommand := fmt.Sprintf("Running authenticated bosh command: update-cloud-config %s/cloud-config.yml (detach: false)", tempDir)
		Expect(actions).To(ContainElement(expectedCommand))
	})

	It("Uploads the concourse stemcell", func() {
		_, _, err := client.Deploy(nil, nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement(ContainSubstring("Running authenticated bosh command: upload-stemcell https://s3.amazonaws.com/bosh-aws-light-stemcells/light-bosh-stemcell-")))
	})

	It("Saves the concourse manifest", func() {
		_, _, err := client.Deploy(nil, nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Saving file to working dir: concourse.yml"))
	})

	It("Deploys concourse", func() {
		_, _, err := client.Deploy(nil, nil, false)
		Expect(err).ToNot(HaveOccurred())
		expectedCommand := fmt.Sprintf("Running authenticated bosh command: --deployment concourse deploy %s/concourse.yml --vars-store %s/concourse-creds.yml --ops-file %s/versions.json --ops-file %s/cup_compatibility.yml --vars-file %s/grafana_dashboard.yml --var deployment_name=concourse --var domain= --var project=happymeal --var web_network_name=public --var worker_network_name=private --var-file postgres_ca_cert=%s/ca.pem --var postgres_host=rds.aws.com --var postgres_port=5432 --var postgres_role=admin --var postgres_password=s3cret --var postgres_host=rds.aws.com --var web_vm_type=concourse-web- --var worker_vm_type=concourse- --var atc_eip=77.77.77.77 (detach: false)", tempDir, tempDir, tempDir, tempDir, tempDir, tempDir)
		Expect(actions).To(ContainElement(expectedCommand))
	})
})
