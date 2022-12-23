package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/allegro/bigcache"
	log "github.com/sirupsen/logrus"
)

// Downloader is the base structure for grabbing the necessary files.
type Downloader struct {
	http.Client
	Waiter sync.WaitGroup
	Types  []string
	Mirror string
	Follow bool
}

var (
	primaryMirror   = os.Getenv("PRIMARY_MIRROR")
	secondaryMirror = os.Getenv("SECONDARY_MIRROR")
)

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

	var buf bytes.Buffer
	req, err := http.NewRequest("GET", rawURL, &buf)
	if err != nil {
		log.WithFields(log.Fields{
			"filename": filename,
			"err":      err,
		}).Error("failed to create request")
	}

	req.Header.Set("User-Agent", "CVDUPDATE/1.0")

	resp, err := d.Do(req)
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

	body, err := io.ReadAll(resp.Body)
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
