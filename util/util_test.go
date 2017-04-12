package util_test

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"bitbucket.org/engineerbetter/concourse-up/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	})
})
