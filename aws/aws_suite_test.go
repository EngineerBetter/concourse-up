package aws_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAws(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aws Suite")
}
