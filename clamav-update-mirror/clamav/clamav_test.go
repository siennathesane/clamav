package clamav_test

import (
	"io/ioutil"
	"os"

	"github.com/pivotal-cloudops/cloudops-ci/concourse/tasks/clamav-update-mirror/clamav"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Clamav", func() {
	Describe("TXT record parse", func() {
		Context("When string format is correct", func() {
			It("returns a version struct", func() {

				expected := clamav.VersionSet{
					Clamav:       "0.99.2",
					Main:         57,
					Daily:        22382,
					Safebrowsing: 45137,
					Bytecode:     283,
				}

				actual, err := clamav.ParseTxtRecordForVersions("0.99.2:57:22382:1476727136:1:63:45137:283")
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})
		})

		Context("When string format is incorrect", func() {
			It("returns an error", func() {
				_, err := clamav.ParseTxtRecordForVersions("0:0:0")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("CVD header parser", func() {
		var (
			tmpFile *os.File
			err     error
		)

		BeforeEach(func() {
			tmpFile, err = ioutil.TempFile("", "test.tvd")
			Expect(err).NotTo(HaveOccurred())

			content := `ClamAV-VDB:23 Jun 2016 11-01 -0400:283:53:63:a3c18c1a448521c27ce915fefbc0d1b9:lYZ9T5viSFqrfnX7qB0ZlxAvpLD0h5XBimF9GimsmJl5Vrlo0U9zSgsRrQvRZ7KSiZXxISggWS+2pQUnfTfDx6mDf+B+boYYXEY+mutHFjOOW7EZ5xIstg3wwju9TaNUlLgIosv3MM6Wyp0wXuze6SqdDnBLw4f3i5k/2x9iaGe:neo:1466694097
`
			err = ioutil.WriteFile(tmpFile.Name(), []byte(content), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err = tmpFile.Close()
			Expect(err).NotTo(HaveOccurred())
			os.RemoveAll(tmpFile.Name())
		})

		It("returns a version number from CVD file", func() {
			actual, err := clamav.ParseCvdVersion(tmpFile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(283))
		})
	})
})
