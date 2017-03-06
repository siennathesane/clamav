package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestClamavUpdateMirror(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClamavUpdateMirror Suite")
}
