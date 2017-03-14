package main

import (
	"github.com/allegro/bigcache"
	"testing"
	"fmt"
	"time"
)

func newTempCache() *bigcache.BigCache {
	c, _ := bigcache.NewBigCache(bigcache.DefaultConfig(time.Second * 30))
	return c
}

func TestDownloader_DownloadFile(t *testing.T) {
	testCache := newTempCache()
	testDL := NewDownloader(false)

	testURL := fmt.Sprintf("%s/%s.cvd", "https://pivotal-clamav-mirror.s3.amazonaws.com", "daily")
	testDL.Waiter.Add(1)
	testDL.DownloadFile(testURL, testCache)

	if _, err := testCache.Get("daily.cvd"); err != nil {
		t.Error(err)
		if err.Error() == "Could not retrieve entry from cache" {
			t.Log("bad download.")
			t.SkipNow()
		}
		t.Fail()
	}
}

func TestDownloader_DownloadDatabase(t *testing.T) {
	testCache := newTempCache()
	testDL := NewDownloader(false)

	testDL.DownloadDatabase(testCache)

	var need = []string{"daily.cvd", "main.cvd", "bytecode.cvd"}
	for file := range need {
		if _, err := testCache.Get(need[file]); err != nil {
			t.Error(err)
			if err.Error() == "Could not retrieve entry from cache" {
				t.Log("bad download.")
				t.SkipNow()
			}
			t.Fail()
		}
	}
}