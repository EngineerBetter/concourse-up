package config_test

import (
	. "github.com/EngineerBetter/concourse-up/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeployArgs", func() {
	var deployArgs *DeployArgs

	BeforeEach(func() {
		deployArgs = &DeployArgs{
			AllowIPs:               "0.0.0.0",
			AWSRegion:              "eu-west-1",
			DBSize:                 "small",
			DBSizeIsSet:            false,
			Domain:                 "",
			GithubAuthClientID:     "",
			GithubAuthClientSecret: "",
			IAAS:        "AWS",
			SelfUpdate:  false,
			TLSCert:     "",
			TLSKey:      "",
			WebSize:     "small",
			WorkerCount: 1,
			WorkerSize:  "xlarge",
		}
	})

	Describe("Validate", func() {
		Context("When the config contains default values", func() {
			It("is valid", func() {
				err := deployArgs.Validate()
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("validateCertFields", func() {
			Context("When all cert fields are specified", func() {
				It("is valid", func() {
					deployArgs.ModifyCerts("ci.engineerbetter.com", "a cool cert", "a cool key")
					err := deployArgs.Validate()
					Expect(err).ToNot(HaveOccurred())
				})
			})
			Context("When TLSKey is not specified", func() {
				It("returns a helpful error", func() {
					deployArgs.ModifyCerts("ci.engineerbetter.com", "a cool cert", "")
					err := deployArgs.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("--tls-cert requires --tls-key to also be provided"))
				})
			})
			Context("When TLSCert is not specified", func() {
				It("returns a helpful error", func() {
					deployArgs.ModifyCerts("cup.engineerbetter.com", "", "a cool key")
					err := deployArgs.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("--tls-key requires --tls-cert to also be provided"))
				})
			})
			Context("When TLSCert and TLSKey are specified but domain is not", func() {
				It("returns a helpful error", func() {
					deployArgs.ModifyCerts("", "a cool cert", "a cool key")
					err := deployArgs.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("custom certificates require --domain to be provided"))
				})
			})
			Context("When just domain is specified", func() {
				It("is valid", func() {
					deployArgs.ModifyCerts("cup.engineerbetter.com", "", "")
					err := deployArgs.Validate()
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
		Context("validateWorkerFields", func() {
			Context("When an invalid number of workers is specified", func() {
				It("returns a helpful error", func() {
					deployArgs.ModifyWorker(0, "xlarge")
					err := deployArgs.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("minimum number of workers is 1"))
				})
			})
			Context("When an invalid worker size is specified", func() {
				It("returns a helpful error", func() {
					deployArgs.ModifyWorker(1, "bananas")
					err := deployArgs.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unknown worker size: `bananas`."))
				})
			})
			Context("When an all worker settings are invalid", func() {
				It("the worker count error takes precedence", func() {
					deployArgs.ModifyWorker(0, "bananas")
					err := deployArgs.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("minimum number of workers is 1"))
				})
			})
		})
		Context("validateWebFields", func() {
			Context("When an invalid web size is specified", func() {
				It("returns a helpful error", func() {
					deployArgs.ModifyWeb("bananas")
					err := deployArgs.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unknown web node size: `bananas`."))
				})
			})
		})
		Context("validateDBFields", func() {
			Context("When an invalid DB size is specified", func() {
				It("returns a helpful error", func() {
					deployArgs.ModifyDB("bananas")
					err := deployArgs.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unknown DB size: `bananas`."))
				})
			})
		})
		Context("validateGithubFields", func() {
			Context("When an all github fields are specified", func() {
				It("is valid", func() {
					deployArgs.ModifyGithub("client ID", "client secret", true)
					err := deployArgs.Validate()
					Expect(err).ToNot(HaveOccurred())
				})
			})
			Context("When only the client ID is specified", func() {
				It("returns a helpful error", func() {
					deployArgs.ModifyGithub("client ID", "", false)
					err := deployArgs.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("--github-auth-client-id requires --github-auth-client-secret to also be provided"))
				})
			})
			Context("When only the client secret is specified", func() {
				It("returns a helpful error", func() {
					deployArgs.ModifyGithub("", "client secret", false)
					err := deployArgs.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("--github-auth-client-secret requires --github-auth-client-id to also be provided"))
				})
			})
		})
	})
})
