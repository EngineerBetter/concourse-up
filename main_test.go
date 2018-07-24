package main_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var (
	cliPath string
)

var _ = Describe("concourse-up", func() {
	var ldFlags []string

	BeforeSuite(func() {
		compilationVars := map[string]string{}

		file, err := os.Open("compilation-vars.json")
		Expect(err).To(Succeed())
		defer file.Close()

		err = json.NewDecoder(file).Decode(&compilationVars)
		Expect(err).To(Succeed())

		ldFlags = []string{
			fmt.Sprintf("-X main.ConcourseUpVersion=%s", "0.0.0"),
		}

		cliPath, err = Build("github.com/EngineerBetter/concourse-up", "-ldflags", strings.Join(ldFlags, " "))
		Expect(err).ToNot(HaveOccurred(), "Error building source")
	})

	AfterSuite(func() {
		CleanupBuildArtifacts()
	})

	It("displays usage instructions on --help", func() {
		command := exec.Command(cliPath, "--help")
		session, err := Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred(), "Error running CLI: "+cliPath)
		Eventually(session).Should(Exit(0))
		Expect(session.Out).To(Say("Concourse-Up - A CLI tool to deploy Concourse CI"))
		Expect(session.Out).To(Say("deploy, d   Deploys or updates a Concourse"))
		Expect(session.Out).To(Say("destroy, x  Destroys a Concourse"))
	})
})
