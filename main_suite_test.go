package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConcourseUp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Concourse-Up Suite")
}
