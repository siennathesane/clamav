package main_test

import (
	"gitlab.apps.prd.central-us-pcf.fnts.io/ops/clamav/clamav-update-mirror/clamav"

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
