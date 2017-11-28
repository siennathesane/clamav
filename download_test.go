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
	"testing"
	"time"

	"github.com/allegro/bigcache"
)

func newTempCache() *bigcache.BigCache {
	c, _ := bigcache.NewBigCache(bigcache.DefaultConfig(time.Second * 1))
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
	if testing.Short() {
		t.Skip("Skipping database test.")
	}
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
