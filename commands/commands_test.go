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
		cliPath, err = Build("github.com/engineerbetter/concourse-up")
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
				Expect(session.Out).To(Say("--region value  AWS region"))
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
})
