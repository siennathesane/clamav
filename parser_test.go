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
	"github.com/allegro/bigcache"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"
)

/*

Raw file header example:

ClamAV-VDB:07 Mar 2017 08-02 -0500:23182:1741572:63:c1537143239006af01e814a4dcd58a48:QC2ZncCPK0AzfYPW8OKvde9GFOO1HyH5qbozl9JZbmlOmZnSV55zWaP9yH9tXiS+JmZWA1277X6pBeTHPCcaqUDakke4W58duZ5mavDGJoWekl3q/5RgVeAg39cM1X4zNf6gER8G+HIWDUka0sRQWal1KXAb1UWkFoKsbHVqgVi:neo:1488891746

Field Definitions:
ClamAV-VDB: Header, defines file type.
07 Mar 2017 08-02 -0500: Creation time.
23182: Version Number
1741572: Number of signatures.
63: Functionality level.
c1...58a48: MD5 checksum.
QC...VqgVi: Digital Signature. Type Unknown.
neo: Builder Name.
Creation time in Epoch Seconds: old file format.

Actual definition:
struct cl_cvd {
char *time;		    2
unsigned int version;   3
unsigned int sigs;	    4
unsigned int fl;	    5
// padding
char *md5;		    /6
char *dsig;		    7
char *builder;	    8
unsigned int stime;	    9
};

*/

const (
	RealHeader = "ClamAV-VDB:07 Mar 2017 08-02 -0500:23182:1741572:63:c1537143239006af01e814a4dcd58a48:QC2ZncCPK0AzfYPW8OKvde9GFOO1HyH5qbozl9JZbmlOmZnSV55zWaP9yH9tXiS+JmZWA1277X6pBeTHPCcaqUDakke4W58duZ5mavDGJoWekl3q/5RgVeAg39cM1X4zNf6gER8G+HIWDUka0sRQWal1KXAb1UWkFoKsbHVqgVi:neo:1488891746                                                                                                                                                                                                                                                 ^_<8B>^H^@^@^@^@^@^B"
)

func newTestRuntime() (*http.Client, *bigcache.BigCache, *sync.WaitGroup) {
	c, _ := bigcache.NewBigCache(bigcache.DefaultConfig(time.Second * 10))
	var wg sync.WaitGroup
	return &http.Client{}, c, &wg
}

func localFile(f string) ([]byte, error) {
	return ioutil.ReadFile(f)
}

func TestParseCvdVersion(t *testing.T) {
	want := 23182

	have, err := ParseCvdVersion([]byte(RealHeader))
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if have != want {
		t.Logf("want %d, have: %d", want, have)
		t.Fail()
	}
}

func TestParseCVD(t *testing.T) {
	cl, c, wg := newTestRuntime()
	url := MainMirror + "/daily.cvd"
	go downloadFile(url, cl, c, "daily", wg)
	wg.Wait()

	testFile, err := c.Get("daily.cvd")
	if err != nil || len(testFile) == 0 {
		t.Error(err)
		testFile, err = localFile("main.cvd")
		if err != nil {
			t.Error("cannot read from cache or local disk! tests cannot run!")
			t.FailNow()
		}
	}

	var errs []error
	_ = ParseCVD(testFile, &errs)
	if len(errs) > 0 {
		t.Errorf("%v", errs)
		t.Fail()
	}
}
