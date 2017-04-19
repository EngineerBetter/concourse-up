package util_test

import (
	"io"

	"bitbucket.org/engineerbetter/concourse-up/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("util functions", func() {
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
