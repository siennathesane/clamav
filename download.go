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
	"github.com/allegro/bigcache"
)

// TODO rewrite most of this at a later time so it's more extensible.

const (
	MainMirror = "http://database.clamav.net"
	// TODO not sure I need this.
	// TextRecord = "current.cvd.clamav.net"
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

// DownloadDatabase downloads the AV definitions and some other basic business logic. It uses the predefined cache to
// save files.
func DownloadDatabase(c *bigcache.BigCache) {
	// TODO add client tracing for InfoSec/auditing.
	var downloadClient = &http.Client{}
	var wg sync.WaitGroup

	// while I should really call an init() function for downloading the AV
	// stuff, this is just easier because I'm being lazy.
	//if _, err := os.Stat(Location); os.IsNotExist(err) {
	//	os.Mkdir(Location, 0600)
	//}

	// added concurrency so it wasn't blocking.
	for _, dbType := range dbTypes {
		wg.Add(1)
		go downloadFile(MainMirror+"/"+dbType+".cvd", downloadClient, c, &wg)

		var cdiffVer int
		cdiffVer = ParseCvdVersion(filepath.Join("/tmp/lib/", dbType+".cvd"))
		cdiffUrl := MainMirror + "/" + dbType + "-" + strconv.Itoa(cdiffVer) + ".cdiff"
		wg.Add(1)
		go downloadFile(cdiffUrl, downloadClient, c, &wg)
	}
	wg.Wait()
	log.Info("done downloading definitions.")
}

//downloadFile performs the download and places it in the /tmp directory.
func downloadFile(rawUrl string, cl *http.Client, c *bigcache.BigCache, wg *sync.WaitGroup) {
	defer wg.Done()

	cvdUrl, err := url.Parse(rawUrl)
	if err != nil {
		log.Errorf("cannot parse %s as url.", rawUrl)
	}

	filename := strings.TrimLeft(cvdUrl.Path, "/")
	log.Println("downloading", filename)

	resp, err := cl.Get(rawUrl)
	if err != nil {
		log.Errorf("failed to download file! %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read body! %s", err)
	}

	// normally this can be deferred, but it needs to closed before adding into the cache.
	if err = resp.Body.Close(); err != nil {
		log.Errorf("cannot close response connection! %s", err)
	}

	if err = c.Set(filename, body); err != nil {
		log.Errorf("cannot add %s to cache! %s", filename, err)
	}
}
