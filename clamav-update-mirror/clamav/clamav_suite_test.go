package clamav_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestClamav(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clamav Suite")
}
