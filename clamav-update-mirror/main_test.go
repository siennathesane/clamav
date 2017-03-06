package main_test

import (
	"github.com/pivotal-cloudops/cloudops-ci/concourse/tasks/clamav-update-mirror/clamav"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {
	Describe("dummy", func() {
		It("does nothing", func() {
			Expect(clamav.Version).To(Equal(1))
		})
	})
})
