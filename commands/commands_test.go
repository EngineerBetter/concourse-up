package commands

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var (
	cliPath string
)

var _ = Describe("commands", func() {
	BeforeSuite(func() {
		var err error
		cliPath, err = Build("github.com/EngineerBetter/concourse-up")
		Expect(err).ToNot(HaveOccurred(), "Error building source")
	})

	AfterSuite(func() {
		CleanupBuildArtifacts()
	})

	Describe("deploy", func() {
		Context("When using --help", func() {
			It("should display usage details", func() {
				command := exec.Command(cliPath, "deploy", "--help")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred(), "Error running CLI: "+cliPath)
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say("concourse-up deploy - Deploys or updates a Concourse"))
				Expect(session.Out).To(Say("--region value"))
				Expect(session.Out).To(Say("--domain value"))
				Expect(session.Out).To(Say("--tls-cert value"))
				Expect(session.Out).To(Say("--tls-key value"))
				Expect(session.Out).To(Say("--db-size value"))
			})
		})

		Context("When no name is passed in", func() {
			It("should display correct usage", func() {
				command := exec.Command(cliPath, "deploy")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Usage is `concourse-up deploy <name>`"))
			})
		})

		Context("When there is a key but no cert", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--domain", "abc.engineerbetter.com", "--tls-key", "-- BEGIN RSA PRIVATE KEY --")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("--tls-key requires --tls-cert to also be provided"))
			})
		})

		Context("When there is a cert but no key", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--domain", "abc.engineerbetter.com", "--tls-cert", "-- BEGIN CERTIFICATE --")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("--tls-cert requires --tls-key to also be provided"))
			})
		})

		Context("When there is a cert and key but no domain", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--tls-key", "-- BEGIN RSA PRIVATE KEY --", "--tls-cert", "-- BEGIN RSA PRIVATE KEY --")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("custom certificates require --domain to be provided"))
			})
		})

		Context("When an invalid worker count is provided", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--workers", "0")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("minimum number of workers is 1"))
			})
		})

		Context("When an invalid worker size is provided", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--worker-size", "small")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("unknown worker size"))
			})
		})

		Context("When an invalid web size is provided", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--web-size", "tiny")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("unknown web node size"))
			})
		})

		Context("When an invalid db size is provided", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--db-size", "huge")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("unknown DB size"))
			})
		})
	})

	Describe("destroy", func() {
		Context("When using --help", func() {
			It("should display usage details", func() {
				command := exec.Command(cliPath, "destroy", "--help")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred(), "Error running CLI: "+cliPath)
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say("concourse-up destroy - Destroys a Concourse"))
			})
		})

		Context("When no name is passed in", func() {
			It("should display correct usage", func() {
				command := exec.Command(cliPath, "destroy")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Usage is `concourse-up destroy <name>`"))
			})
		})
	})

	Describe("info", func() {
		Context("When using --help", func() {
			It("should display usage details", func() {
				command := exec.Command(cliPath, "info", "--help")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred(), "Error running CLI: "+cliPath)
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say("concourse-up info - Fetches information on a deployed environment"))
			})
		})

		Context("When no name is passed in", func() {
			It("should display correct usage", func() {
				command := exec.Command(cliPath, "info")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Usage is `concourse-up info <name>`"))
			})
		})
	})
})
