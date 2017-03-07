package main

import (
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"net/http"
	"io/ioutil"
	"sync"
	log "github.com/Sirupsen/logrus"
	"bufio"
)

// TODO rewrite most of this at a later time so it's more extensible.

const (
	MainMirror = "http://database.lib.net"
	Location   = "/tmp/clamv"
)

var dbTypes = []string{"main", "bytecode", "daily"}

// ParseCvdVersion reads a ClamAV CVD file and parses it for the version.
func ParseCvdVersion(cvdFile string) int {
	fh, err := os.Open(cvdFile)
	if err != nil {
		log.Error(err)
		return 0
	}
	defer fh.Close()

	head := make([]byte, 512)
	_, err = fh.Read(head)
	if err != nil {
		log.Errorf("cannot read %s", cvdFile)
		return 0
	}

	headStr := string(head)
	headParts := strings.Split(headStr, ":")
	if len(headParts) < 3 {
		log.Errorf("invalid header string: %s", headStr)
		return 0
	}

	var verNum int
	verNum, err = strconv.Atoi(headParts[2])
	if err != nil {
		log.Error(err)
		return 0
	}

	return verNum
}

// DownloadDatabase downloads the AV definitions and some other basic business logic.
func DownloadDatabase() {
	// TODO add client tracing for InfoSec/auditing.
	var downloadClient = &http.Client{}
	var wg sync.WaitGroup

	// while I should really call an init() function for downloading the AV
	// stuff, this is just easier because I'm being lazy.
	if _, err := os.Stat(Location); os.IsNotExist(err) {
		os.Mkdir(Location, 0600)
	}

	// added concurrency so it wasn't blocking.
	for _, dbType := range dbTypes {
		wg.Add(1)
		go downloadFile(MainMirror+"/"+dbType+".cvd", downloadClient, &wg)

		var cdiffVer int
		cdiffVer = ParseCvdVersion(filepath.Join("/tmp/lib/", dbType+".cvd"))
		cdiffUrl := MainMirror + "/" + dbType + "-" + strconv.Itoa(cdiffVer) + ".cdiff"
		wg.Add(1)
		go downloadFile(cdiffUrl, downloadClient, &wg)
	}
	wg.Wait()
}

//downloadFile performs the download and places it in the /tmp directory.
func downloadFile(rawUrl string, cl *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	var err error
	var cvdUrl *url.URL
	if cvdUrl, err = url.Parse(rawUrl); err != nil {
		log.Errorf("cannot parse %s as url.", rawUrl)
	}

	filename := strings.TrimLeft(cvdUrl.Path, "/")
	log.Println("downloading", filename)

	cvdDest := filepath.Join(Location, filename)
	resp, err := cl.Get(rawUrl)
	if err != nil {
		log.Errorf("failed to download file! %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read body! %s", err)
	}

	fh, err := os.OpenFile(cvdDest, os.O_TRUNC, 0666)
	if err != nil {
		log.Error(err)
	}

	writer := bufio.NewWriter(fh)
	defer fh.Close()

	_, err = fh.Write(body)
	if err != nil {
		log.Error(err)
	}
	writer.Flush()
}
