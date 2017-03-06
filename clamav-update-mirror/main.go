package main

import (
	"log"
	"os"

	"github.com/pivotal-cloudops/cloudops-ci/concourse/tasks/clamav-update-mirror/clamav"
)

func main() {

	dbDir := os.Args[1]

	clamav.DownloadDatabase(dbDir)
	log.Println("Done downloading")
}
