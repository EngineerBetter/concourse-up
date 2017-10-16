package bosh_test

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	. "github.com/EngineerBetter/concourse-up/bosh"
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

	fakeDBRunner := func(sql string) error {
		actions = append(actions, fmt.Sprintf("Running SQL: %s", sql))
		return nil
	}

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
			PublicKey:        "example-public-key",
			PrivateKey:       "example-private-key",
			Region:           "eu-west-1",
			Deployment:       "concourse-up-happymeal",
			Project:          "happymeal",
			TFStatePath:      "example-path",
			DirectorUsername: "admin",
			DirectorPassword: "secret123",
			ConcourseDBName:  "concourse_atc",
			RDSUsername:      "admin",
			RDSPassword:      "s3cret",
		}

		client = NewClient(
			exampleConfig,
			terraformMetadata,
			directorClient,
			fakeDBRunner,
			bytes.NewBuffer(nil),
			bytes.NewBuffer(nil),
		)
	})

	Context("When an initial director state exists", func() {
		It("Saves the director state", func() {
			stateFileBytes := []byte("{}")
			_, err := client.Deploy(stateFileBytes, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(actions).To(ContainElement("Saving file to working dir: director-state.json"))
		})
	})

	Context("When an initial director state does not exist", func() {
		It("Does not save the director state", func() {
			_, err := client.Deploy(nil, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(actions).ToNot(ContainElement("Saving file to working dir: director-state.json"))
		})
	})

	It("Saves the private key", func() {
		_, err := client.Deploy(nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Saving file to working dir: director.pem"))
	})

	It("Saves the manifest", func() {
		_, err := client.Deploy(nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Saving file to working dir: director.yml"))
	})

	It("Deploys the director", func() {
		_, err := client.Deploy(nil, false)
		Expect(err).ToNot(HaveOccurred())
		expectedCommand := fmt.Sprintf("Running bosh command: create-env %s/director.yml --state %s/director-state.json", tempDir, tempDir)
		Expect(actions).To(ContainElement(expectedCommand))
	})

	It("Saves the cloud config", func() {
		_, err := client.Deploy(nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Saving file to working dir: cloud-config.yml"))
	})

	It("Updates the cloud config", func() {
		_, err := client.Deploy(nil, false)
		Expect(err).ToNot(HaveOccurred())
		expectedCommand := fmt.Sprintf("Running authenticated bosh command: update-cloud-config %s/cloud-config.yml (detach: false)", tempDir)
		Expect(actions).To(ContainElement(expectedCommand))
	})

	It("Uploads the concourse stemcell", func() {
		_, err := client.Deploy(nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-stemcell COMPILE_TIME_VARIABLE_bosh_concourseStemcellURL (detach: false)"))
	})

	It("Uploads the each bosh release", func() {
		_, err := client.Deploy(nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-release COMPILE_TIME_VARIABLE_bosh_concourseReleaseURL --stemcell ubuntu-trusty/COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion (detach: false)"))
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-release COMPILE_TIME_VARIABLE_bosh_grafanaReleaseURL --stemcell ubuntu-trusty/COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion (detach: false)"))
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-release COMPILE_TIME_VARIABLE_bosh_gardenReleaseURL --stemcell ubuntu-trusty/COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion (detach: false)"))
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-release COMPILE_TIME_VARIABLE_bosh_influxDBReleaseURL --stemcell ubuntu-trusty/COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion (detach: false)"))
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-release COMPILE_TIME_VARIABLE_bosh_riemannReleaseURL --stemcell ubuntu-trusty/COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion (detach: false)"))
	})

	It("Saves the concourse manifest", func() {
		_, err := client.Deploy(nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Saving file to working dir: concourse.yml"))
	})

	It("Creates the default concourse DB", func() {
		_, err := client.Deploy(nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Running SQL: CREATE DATABASE concourse_atc;"))
	})

	It("Deploys concourse", func() {
		_, err := client.Deploy(nil, false)
		Expect(err).ToNot(HaveOccurred())
		expectedCommand := fmt.Sprintf("Running authenticated bosh command: --deployment concourse deploy %s/concourse.yml (detach: false)", tempDir)
		Expect(actions).To(ContainElement(expectedCommand))
	})

})
