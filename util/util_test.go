package util_test

import (
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"bitbucket.org/engineerbetter/concourse-up/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("util functions", func() {

	var path string
	usr, _ := user.Current()

	Describe("file functions", func() {
		Describe("Path", func() {
			Context("When the path begins with `~`", func() {
				JustBeforeEach(func() {
					path = "~/.concourse-up"
				})

				expectedPath := filepath.Join(usr.HomeDir, ".concourse-up")
				It("should substitute `~` for the current user's home directory", func() {
					path, err := util.Path(path)
					Ω(err).Should(BeNil())
					Ω(path).Should(Equal(expectedPath))
				})
			})

			Context("When the path contains `~` other than at the beginning", func() {
				JustBeforeEach(func() {
					path = "foobar/~/.concourse-up"
				})

				It("should return an error", func() {
					_, err := util.Path(path)
					Ω(err).Should(MatchError("Invalid Path"))
				})
			})

			Context("When the path does not contain `~`", func() {
				JustBeforeEach(func() {
					path = "foobar/.concourse-up"
				})
				It("should use the path as written", func() {
					path, err := util.Path(path)
					Ω(err).Should(BeNil())
					Ω(path).Should(Equal(path))
				})
			})
		})

		Describe("AssertFileExists", func() {
			Context("When the file does not already exist", func() {
				var tmpdir string
				JustBeforeEach(func() {
					tmpdir, err := ioutil.TempDir("", "example")
					util.CheckErr(err)
					path = filepath.Join(tmpdir, "tempfile")
				})

				AfterEach(func() {
					os.RemoveAll(tmpdir) // clean up
				})

				It("should create the file in the location given by `path`", func() {
					Ω(path).ShouldNot(BeAnExistingFile())
					err := util.AssertFileExists(path)
					Ω(err).Should(BeNil())
					Ω(path).Should(BeAnExistingFile())
				})
			})

			Context("When the file exists", func() {
				JustBeforeEach(func() {
					tmpfile, err := ioutil.TempFile("", "example")
					util.CheckErr(err)
					path = tmpfile.Name()

					text := []byte("hello world")
					err = ioutil.WriteFile(path, text, 0600)
					util.CheckErr(err)
				})

				AfterEach(func() {
					os.Remove(path) // clean up
				})

				It("should not overwrite the file at `path`", func() {
					Ω(path).Should(BeAnExistingFile())
					err := util.AssertFileExists(path)
					Ω(err).Should(BeNil())
					Ω(path).Should(BeAnExistingFile())
					contents, err := ioutil.ReadFile(path)
					Ω(string(contents)).Should(Equal("hello world"))
				})
			})
		})

		Describe("AssertDirExists", func() {
			Context("When the dir does not already exist", func() {
				var homeDir string
				JustBeforeEach(func() {
					homeDir, err := ioutil.TempDir("", "example")
					util.CheckErr(err)
					path = filepath.Join(homeDir, ".concourse-up")
				})

				AfterEach(func() {
					os.RemoveAll(homeDir) // clean up
				})

				It("should create the file in the location given by `path`", func() {
					_, err := os.Stat(path)
					Ω(os.IsNotExist(err)).Should(BeTrue())
					err = util.AssertDirExists(path)
					Ω(err).Should(BeNil())
					_, err = os.Stat(path)
					Ω(err).Should(BeNil())
				})
			})

			Context("When the dir exists", func() {
				var homeDir string
				var dirPath string
				JustBeforeEach(func() {
					homeDir, err := ioutil.TempDir("", "example")
					util.CheckErr(err)
					dirPath = filepath.Join(homeDir, ".concourse-up")
					err = os.MkdirAll(dirPath, 0755)
					util.CheckErr(err)
					tmpfile, err := ioutil.TempFile(dirPath, "example")
					util.CheckErr(err)
					path = tmpfile.Name()

					text := []byte("hello world")
					err = ioutil.WriteFile(path, text, 0644)
					util.CheckErr(err)
				})

				AfterEach(func() {
					os.RemoveAll(homeDir) // clean up
				})

				It("should not modify the file in `path`", func() {
					_, err := os.Stat(path)
					Ω(os.IsNotExist(err)).Should(BeFalse())
					Ω(path).Should(BeAnExistingFile())
					err = util.AssertDirExists(dirPath)
					Ω(err).Should(BeNil())
					Ω(path).Should(BeAnExistingFile())
					contents, err := ioutil.ReadFile(path)
					Ω(string(contents)).Should(Equal("hello world"))
				})
			})
		})
	})

	Describe("confirmation check", func() {
		var stdin io.ReadWriter
		var stdout io.Writer

		BeforeEach(func() {
			stdin = gbytes.NewBuffer()
			stdout = gbytes.NewBuffer()
		})

		Context("When the user responds with 'yes'", func() {
			It("Returns true", func() {
				stdin.Write([]byte("yes\n"))
				returnVal, err := util.CheckConfirmation(stdin, stdout, "serrano")
				Eventually(stdout).Should(gbytes.Say(`Are you sure you want to destroy serrano?`))
				Eventually(returnVal).Should(Equal(true))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("When the user responds with no", func() {
			It("Returns false", func() {
				stdin.Write([]byte("no\n"))
				returnVal, err := util.CheckConfirmation(stdin, stdout, "serrano")
				Eventually(stdout).Should(gbytes.Say(`Are you sure you want to destroy serrano?`))
				Eventually(returnVal).Should(Equal(false))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("When the user responds with any thing else", func() {
			It("Returns a meaningful error message", func() {
				stdin.Write([]byte("what\n"))
				returnVal, err := util.CheckConfirmation(stdin, stdout, "serrano")
				Eventually(stdout).Should(gbytes.Say(`Are you sure you want to destroy serrano?`))
				Eventually(returnVal).Should(Equal(false))
				Expect(err).To(MatchError("Input not recognized: `what`"))
			})
		})
	})
})
