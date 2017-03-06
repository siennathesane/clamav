package main

import (
	"log"
	"os"
	"./clamav"
)


func main() {

	dbDir := os.Args[1]

	clamav.DownloadDatabase(dbDir)
	log.Println("Done downloading")
}
