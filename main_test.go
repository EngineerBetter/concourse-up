package main_test

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

var _ = Describe("concourse-up", func() {
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
	args = []string{"--help"}
	It("displays usage instructions on --help", func() {
		Eventually(session).Should(Exit(0))
		Ω(session.Out).Should(Say("Concourse-Up - A CLI tool to deploy Concourse CI"))
		Ω(session.Out).Should(Say("deploy, d  Deploys or updates a Concourse"))
	})
})
