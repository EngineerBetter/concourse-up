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
	args    []string
	session *Session
	err     error
)

var _ = Describe("commands", func() {
	BeforeSuite(func() {
		cliPath, err = Build("bitbucket.org/engineerbetter/concourse-up")
		Ω(err).ShouldNot(HaveOccurred(), "Error building source")
	})

	AfterSuite(func() {
		CleanupBuildArtifacts()
	})

	JustBeforeEach(func() {
		command := exec.Command(cliPath, args...)
		session, err = Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred(), "Error running CLI: "+cliPath)
	})

	Describe("deploy", func() {
		Context("When using --help", func() {
			BeforeEach(func() {
				args = []string{"deploy", "--help"}
			})
			It("should display usage details", func() {
				Eventually(session).Should(Exit(0))
				Ω(session.Out).Should(Say("concourse-up deploy - Deploys or updates a Concourse"))
				Ω(session.Out).Should(Say("--region value  AWS region"))
			})
		})
		Context("When no name is passed in", func() {
			BeforeEach(func() {
				args = []string{"deploy"}
			})

			It("should display correct usage", func() {
				Eventually(session).Should(Exit(0))
				Ω(session.Out).Should(Say("Usage is `concourse-up deploy <name>`"))
			})
		})
	})
})
