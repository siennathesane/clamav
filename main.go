package main

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/robfig/cron.v2"
	"net/http"
	"github.com/cloudfoundry-community/go-cfenv"
	"fmt"
)

const (
	DefaultPort = 8080
)

func init() {
	// TODO add runtime.Caller(1) info to it.
	log.SetFormatter(&log.JSONFormatter{})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w,r)
		log.WithFields(log.Fields{
			"host": r.Host,
		})
	})
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

	log.Infof("starting server and initial seed.")
	DownloadDatabase()

	// start a new crontab asynchronously.
	c := cron.New()
	c.AddFunc("@every 3h", func() { DownloadDatabase() })
	c.Start()

	log.Info("done seeding and started cron job for definition downloads.")

	log.Fatal(http.ListenAndServe(port, logMiddleware(http.FileServer(http.Dir(Location)))))
}
