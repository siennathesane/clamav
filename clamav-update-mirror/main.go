package main

import (
	"log"
	"os"

	"gitlab.apps.prd.central-us-pcf.fnts.io/ops/clamav/clamav-update-mirror/clamav"
)

func main() {

	dbDir := os.Args[1]

	clamav.DownloadDatabase(dbDir)
	log.Println("Done downloading")
}
