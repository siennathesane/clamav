/*
   Copyright 2017 Mike Lloyd

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/allegro/bigcache"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TODO rewrite most of this at a later time so it's more extensible.

const (
	// Primary mirror for ClamAV definitions.
	primaryMirror = "http://database.clamav.net"

	// TODO not sure I need this.
	// TextRecord = "current.cvd.clamav.net"
)

var dbTypes = []string{"main", "bytecode", "daily"}

// DownloadDatabase downloads the AV definitions and some other basic business logic. It uses the predefined cache to
// save files.
func DownloadDatabase(c *bigcache.BigCache) {
	// TODO add client tracing for InfoSec/auditing.
	var downloadClient = &http.Client{
		Timeout: time.Duration(time.Second * 5),
	}
	var wg sync.WaitGroup

	// added concurrency so it wasn't blocking.
	for _, dbType := range dbTypes {
		wg.Add(1)
		go downloadFile(primaryMirror+"/"+dbType+".cvd", downloadClient, c, dbType, true, &wg)
	}
	wg.Wait()
	log.Info("done downloading definitions.")
}

// downloadFile performs the download and places it in the /tmp directory.
func downloadFile(rawURL string, cl *http.Client, c *bigcache.BigCache, dbType string, follow bool, wg *sync.WaitGroup) {
	defer wg.Done()

	cvdURL, err := url.Parse(rawURL)
	if err != nil {
		log.WithFields(log.Fields{
			"url": cvdURL,
		}).Error("cannot parse url.")
	}

	filename := strings.TrimLeft(cvdURL.Path, "/")
	log.WithFields(log.Fields{
		"filename": filename,
	}).Info("downloading definition.")

	resp, err := cl.Get(rawURL)
	if err != nil {
		log.WithFields(log.Fields{
			"filename": filename,
			"err":      err,
		}).Error("failed to download file!")
	}

	if resp.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{
			"status_code": resp.StatusCode,
			"filename":    filename,
		}).Error("problem downloading remote definition!")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithField("err", err).Errorf("failed to read body!")
	}

	// normally this can be deferred, but it needs to closed before adding into the cache.
	if err = resp.Body.Close(); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorf("cannot close response connection!")
	}

	if err = c.Set(filename, body); err != nil {
		log.Errorf("cannot add %s to cache! %s", filename, err)
	}

	log.WithFields(log.Fields{
		"filename": filename,
	}).Info("added to cache!")

	cdiffVer, err := ParseCvdVersion(body)
	if err == nil && follow {
		cdiffURL := primaryMirror + "/" + dbType + "-" + strconv.Itoa(cdiffVer) + ".cdiff"
		log.WithField("url", cdiffURL).Debug(cdiffURL)
		wg.Add(1)
		go downloadFile(cdiffURL, cl, c, dbType, follow, wg)
	}

}

// ParseCvdVersion reads a ClamAV CVD file and parses it for the version.
func ParseCvdVersion(cvd []byte) (int, error) {
	var header []byte
	header = append(header, cvd[0:512]...)

	headStr := string(header)
	headParts := strings.Split(headStr, ":")
	if len(headParts) < 3 {
		log.WithFields(log.Fields{
			"err": errors.New("bad def header"),
		}).Error("invalid header string")
		return 0, errors.New("bad def header")
	}

	verNum, err := strconv.Atoi(headParts[2])
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("invalid header string")
		return 0, err
	}

	return verNum, nil
}
