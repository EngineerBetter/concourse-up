package bosh_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	. "github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete", func() {
	var actions []string
	var tempDir string
	var createEnvOutput string

	directorClient := &FakeDirectorClient{
		FakeRunCommand: func(stdout, stderr io.Writer, args ...string) error {
			actions = append(actions, fmt.Sprintf("Running bosh command: %s", strings.Join(args, " ")))

			err := ioutil.WriteFile(filepath.Join(tempDir, "director-state.json"), []byte("{ some state }"), 0700)
			Expect(err).ToNot(HaveOccurred())

			if strings.Contains(strings.Join(args, " "), "delete-env") {
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
		createEnvOutput = "Finished deleting deployment"

		var err error
		tempDir, err = ioutil.TempDir("", "bosh_test")
		Expect(err).ToNot(HaveOccurred())

		terraformMetadata := &terraform.Metadata{
			ATCPublicIP:              terraform.MetadataStringValue{Value: "77.77.77.77"},
			ATCSecurityGroupID:       terraform.MetadataStringValue{Value: "sg-999"},
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
		q := mock.ExpectPrepare("CREATE DATABASE \\$1")
		q.ExpectExec().WithArgs("concourse_atc").WillReturnError(errors.New(`pq: database "concourse_atc" already exists`))
		q.ExpectExec().WithArgs("uaa").WillReturnError(errors.New(`pq: database "uaa" already exists`))
		q.ExpectExec().WithArgs("credhub").WillReturnError(errors.New(`pq: database "credhub" already exists`))

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
			_, err := client.Delete(stateFileBytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(actions).To(ContainElement("Saving file to working dir: director-state.json"))
		})
	})

	Context("When an initial director state does not exist", func() {
		It("Does not save the director state", func() {
			_, err := client.Delete(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(actions).ToNot(ContainElement("Saving file to working dir: director-state.json"))
		})
	})

	It("Saves the private key", func() {
		_, err := client.Delete(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Saving file to working dir: director.pem"))
	})

	It("Deletes the director", func() {
		_, err := client.Delete(nil)
		Expect(err).ToNot(HaveOccurred())
		expectedCommand := fmt.Sprintf("Running bosh command: delete-env %s/director.yml --state %s/director-state.json", tempDir, tempDir)
		Expect(actions).To(ContainElement(expectedCommand))
	})

	It("Deletes concourse", func() {
		_, err := client.Delete(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(actions).To(ContainElement("Running authenticated bosh command: --deployment concourse delete-deployment --force (detach: false)"))
	})

})
