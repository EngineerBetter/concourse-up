package bosh_test

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
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

		db, mock, err := sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
		mock.ExpectExec("CREATE DATABASE concourse_atc").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("CREATE DATABASE uaa").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("CREATE DATABASE credhub").WillReturnResult(sqlmock.NewResult(0, 0))
		client = NewClient(
			exampleConfig,
			terraformMetadata,
			directorClient,
			db,
			bytes.NewBuffer(nil),
			bytes.NewBuffer(nil),
		)
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
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-stemcell COMPILE_TIME_VARIABLE_bosh_concourseStemcellURL (detach: false)"))
	})

	It("Uploads the each bosh release", func() {
		_, _, err := client.Deploy(nil, nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-release --stemcell ubuntu-trusty/COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion COMPILE_TIME_VARIABLE_bosh_concourseReleaseURL (detach: false)"))
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-release --stemcell ubuntu-trusty/COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion COMPILE_TIME_VARIABLE_bosh_grafanaReleaseURL (detach: false)"))
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-release --stemcell ubuntu-trusty/COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion COMPILE_TIME_VARIABLE_bosh_gardenReleaseURL (detach: false)"))
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-release --stemcell ubuntu-trusty/COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion COMPILE_TIME_VARIABLE_bosh_influxDBReleaseURL (detach: false)"))
		Expect(actions).To(ContainElement("Running authenticated bosh command: upload-release --stemcell ubuntu-trusty/COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion COMPILE_TIME_VARIABLE_bosh_riemannReleaseURL (detach: false)"))
	})

	It("Saves the concourse manifest", func() {
		_, _, err := client.Deploy(nil, nil, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Saving file to working dir: concourse.yml"))
	})

	It("Deploys concourse", func() {
		_, _, err := client.Deploy(nil, nil, false)
		Expect(err).ToNot(HaveOccurred())
		expectedCommand := fmt.Sprintf("Running authenticated bosh command: --deployment concourse deploy %s/concourse.yml --vars-store %s/concourse-creds.yml (detach: false)", tempDir, tempDir)
		Expect(actions).To(ContainElement(expectedCommand))
	})

})
