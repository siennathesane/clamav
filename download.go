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
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/allegro/bigcache"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	// Mirrors for ClamAV definitions. The standard mirror is slow as all get out.
	secondaryMirror = "http://database.clamav.net"
	primaryMirror   = "https://pivotal-clamav-mirror.s3.amazonaws.com"
)

// Downloader is the base structure for grabbing the necessary files.
type Downloader struct {
	http.Client
	Waiter sync.WaitGroup
	Types  []string
	Mirror string
	Follow bool
}

// NewDownloader will create a new download client which manages the CVD files.
func NewDownloader(f bool) *Downloader {
	var tempClient http.Client
	_, err := tempClient.Get(primaryMirror + "/bytecode.cvd")
	if err != nil {
		return &Downloader{
			Types: []string{
				"main",
				"bytecode",
				"daily",
			},
			Mirror: secondaryMirror,
			Follow: f,
		}
	} else {
		return &Downloader{
			Types: []string{
				"main",
				"bytecode",
				"daily",
			},
			Mirror: primaryMirror,
			Follow: f,
		}
	}
}

// DownloadDatabase downloads the AV definitions and some other basic business logic. It uses the predefined cache to
// save files.
func (d *Downloader) DownloadDatabase(c *bigcache.BigCache) {
	for idx := range d.Types {
		d.Waiter.Add(1)
		rawURL := fmt.Sprintf("%s/%s.cvd", primaryMirror, d.Types[idx])
		go d.DownloadFile(rawURL, c)
	}
	d.Waiter.Wait()

	log.Info("done downloading definitions.")
}

func (d *Downloader) CDiffHelper(s string, i int) string {
	return fmt.Sprintf("%s/%s-%d.cdiff", d.Mirror, s, i)
}

// downloadFile performs the download and places it in the /tmp directory.
func (d *Downloader) DownloadFile(rawURL string, c *bigcache.BigCache) {
	defer d.Waiter.Done()

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

	resp, err := d.Get(rawURL)
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
		}).Error("problem downloading remote definition")
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

	// parse the CVD and make sure it's valid, otherwise, exit.
	var errs []error
	avDefs := ParseCVD(body, &errs)
	if len(errs) > 0 {
		for err := range errs {
			log.WithFields(log.Fields{
				"type":     "parsing error",
				"filename": filename,
			}).Error(err)
		}
		return
	}

	if !avDefs.Header.MD5Valid {
		log.WithField("filename", filename).Error("the md5 is not valid, will not add to cache.")
		return
	}

	if err = c.Set(filename, body); err != nil {
		log.Errorf("cannot add %s to cache! %s", filename, err)
	}

	log.WithFields(log.Fields{
		"filename": filename,
	}).Info("added to cache!")

	if d.Follow {
		cDiffURL := d.CDiffHelper(filename, int(avDefs.Header.Version))
		log.WithField("url", cDiffURL).Debug(cDiffURL)
		d.Waiter.Add(1)
		go d.DownloadFile(cDiffURL, c)
	}
}
