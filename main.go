package main

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/robfig/cron.v2"
	"net/http"
	"github.com/cloudfoundry-community/go-cfenv"
	"fmt"
	"github.com/allegro/bigcache"
	"time"
)

const (
	DefaultPort = 8080
)

func init() {
	// TODO add runtime.Caller(1) info to it.
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	var port string

	// this logic just feels weird to me. idk.
	appEnv, err := cfenv.Current()
	if err != nil {
		log.Error(err)
		port = fmt.Sprintf(":%d", DefaultPort)
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
		log.Errorf("cannot initialise cache. %s")
	}

	log.Infof("starting server and initial seed.")
	DownloadDatabase(cache)

	// start a new crontab asynchronously.
	c := cron.New()
	c.AddFunc("@every 3h", func() { DownloadDatabase(cache) })
	c.Start()

	log.Info("started cron job for definition updates.")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cacheHandler(w, r, cache)
	})

	log.Fatal(http.ListenAndServe(port, nil))
}

// cacheHandler is just a standard handler, but returns stuff from cache.
func cacheHandler(w http.ResponseWriter, r *http.Request, c *bigcache.BigCache) {
	entry, err := c.Get(r.URL.Path[1:])
	if err != nil {
		log.Error(err)
	}

	log.Infof("looking in cache for: %s", r.URL.Path[1:])
	w.Write(entry)
}
