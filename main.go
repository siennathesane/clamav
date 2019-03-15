package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/allegro/bigcache"
	"github.com/cloudfoundry-community/go-cfenv"
	log "github.com/sirupsen/logrus"
	"gopkg.in/robfig/cron.v2"
)

var (
	defaultPort = 8080
)

func init() {
	// TODO add runtime.Caller(1) info to it.
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	flag.IntVar(&defaultPort, "port", 8080, "Port to run the server on.")
}

func main() {
	flag.Parse()
	var port string

	// this logic just feels weird to me. idk.
	appEnv, err := cfenv.Current()
	if err != nil {
		log.Error(err)
		port = fmt.Sprintf(":%d", defaultPort)
	} else {
		port = fmt.Sprintf(":%d", appEnv.Port)
	}

	// overkill, but it's a sane library.
	// we're going to cache the AV definition files.
	cache, err := bigcache.NewBigCache(bigcache.Config{
		MaxEntrySize:       500,
		Shards:             1024,
		LifeWindow:         time.Hour * 3,
		MaxEntriesInWindow: 1000 * 10 * 60,
		Verbose:            true,
		HardMaxCacheSize:   0,
	})

	if err != nil {
		log.Errorf("cannot initialise cache. %s", err)
	}

	// let the initial seed run in the background so the web server can start.
	log.Info("starting initial seed in the background.")
	dl := NewDownloader(true)
	dl.DownloadDatabase(cache)

	// start a new crontab asynchronously.
	c := cron.New()
	if _, err = c.AddFunc("@hourly", func() { NewDownloader(true).DownloadDatabase(cache) }); err != nil {
		log.Error(err)
	}
	c.Start()

	log.Info("started cron job for definition updates.")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cacheHandler(w, r, cache)
	})

	log.Fatal(http.ListenAndServe(port, nil))
}

// cacheHandler is just a standard handler, but returns stuff from cache.
func cacheHandler(w http.ResponseWriter, r *http.Request, c *bigcache.BigCache) {
	filename := r.URL.Path[1:]

	// logs from the gorouter.
	if strings.Contains(filename, "cloudfoundry") {
		log.Warn("nothing to see here, move along.")
		http.NotFound(w, r)
	}

	entry, err := c.Get(filename)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"filename": filename,
		}).Error("cannot return cached file!")
		log.Error(err)
		http.NotFound(w, r)
	}

	log.WithFields(log.Fields{
		"filename": filename,
	}).Info("found!")
	w.Write(entry)
}
