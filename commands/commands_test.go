package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
		compilationVars := map[string]string{}

		file, err := os.Open(filepath.Join("..", "compilation-vars.json"))
		Expect(err).To(Succeed())
		defer file.Close()

		err = json.NewDecoder(file).Decode(&compilationVars)
		Expect(err).To(Succeed())

		ldflags := []string{
			fmt.Sprintf("-X main.ConcourseUpVersion=%s", "0.0.0"),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.ConcourseStemcellURL=%s", compilationVars["concourse_stemcell_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.ConcourseStemcellVersion=%s", compilationVars["concourse_stemcell_version"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.ConcourseStemcellSHA1=%s", compilationVars["concourse_stemcell_sha1"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.ConcourseReleaseURL=%s", compilationVars["concourse_release_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.ConcourseReleaseVersion=%s", compilationVars["concourse_release_version"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.ConcourseReleaseSHA1=%s", compilationVars["concourse_release_sha1"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.RiemannReleaseURL=%s", compilationVars["riemann_release_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.RiemannReleaseVersion=%s", compilationVars["riemann_release_version"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.RiemannReleaseSHA1=%s", compilationVars["riemann_release_sha1"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.GrafanaReleaseURL=%s", compilationVars["grafana_release_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.GrafanaReleaseVersion=%s", compilationVars["grafana_release_version"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.GrafanaReleaseSHA1=%s", compilationVars["grafana_release_sha1"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.InfluxDBReleaseURL=%s", compilationVars["influxdb_release_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.InfluxDBReleaseVersion=%s", compilationVars["influxdb_release_version"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.InfluxDBReleaseSHA1=%s", compilationVars["influxdb_release_sha1"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.GardenReleaseURL=%s", compilationVars["garden_release_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.GardenReleaseVersion=%s", compilationVars["garden_release_version"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.GardenReleaseSHA1=%s", compilationVars["garden_release_sha1"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.DirectorStemcellURL=%s", compilationVars["director_stemcell_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.DirectorStemcellSHA1=%s", compilationVars["director_stemcell_sha1"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.DirectorStemcellVersion=%s", compilationVars["director_stemcell_version"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.DirectorCPIReleaseURL=%s", compilationVars["director_bosh_cpi_release_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.DirectorCPIReleaseVersion=%s", compilationVars["director_bosh_cpi_release_version"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.DirectorCPIReleaseSHA1=%s", compilationVars["director_bosh_cpi_release_sha1"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.DirectorReleaseURL=%s", compilationVars["director_bosh_release_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.DirectorReleaseVersion=%s", compilationVars["director_bosh_release_version"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/bosh.DirectorReleaseSHA1=%s", compilationVars["director_bosh_release_sha1"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/fly.DarwinBinaryURL=%s", compilationVars["fly_darwin_binary_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/fly.LinuxBinaryURL=%s", compilationVars["fly_linux_binary_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/fly.WindowsBinaryURL=%s", compilationVars["fly_windows_binary_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/director.DarwinBinaryURL=%s", compilationVars["director_darwin_binary_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/director.LinuxBinaryURL=%s", compilationVars["director_linux_binary_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/director.WindowsBinaryURL=%s", compilationVars["director_windows_binary_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/terraform.DarwinBinaryURL=%s", compilationVars["terraform_darwin_binary_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/terraform.LinuxBinaryURL=%s", compilationVars["terraform_linux_binary_url"]),
			fmt.Sprintf("-X github.com/EngineerBetter/concourse-up/terraform.WindowsBinaryURL=%s", compilationVars["terraform_windows_binary_url"]),
		}

		cliPath, err = Build("github.com/EngineerBetter/concourse-up", "-ldflags", strings.Join(ldflags, " "))
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
				Expect(session.Err).To(Say("minimum of workers is 1"))
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
