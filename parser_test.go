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
	"io/ioutil"
	"strings"
	"testing"
)

func newLocalFile(f string) ([]byte, error) {
	return ioutil.ReadFile(f)
}

func newSplit(b []byte) ([]byte, []byte) {
	var header []byte
	var def []byte
	header = append(header, b[0:headerLength]...)
	def = append(def, b[headerLength:]...)
	return header, def
}

func TestParseCVD(t *testing.T) {
	testFile, err := newLocalFile("filedefs/bytecode.cvd")
	if err != nil {
		t.Error("cannot read from cache or local disk! tests cannot run!")
		t.FailNow()
	}

	var errs []error
	testRes := ParseCVD(testFile, &errs)
	if len(errs) > 0 {
		t.Errorf("%v", errs)
		t.Fail()
	}
	t.Logf("CVD Header: %#v", testRes)
}

func TestHeaderFields_ParseTime(t *testing.T) {
	testTime := "09 Mar 2017 16-12 -0500"
	testHeader := newEmptyHeader()
	testHeader.ParseTime(testTime)
	if len(testHeader.Problems) > 0 {
		t.Error(testHeader.Problems)
		t.Fail()
	}
}

func TestHeaderFields_Atou(t *testing.T) {
	have := "1234"
	want := uint(1234)

	testHeader := newEmptyHeader()
	got := testHeader.Atou(have)

	if want != got {
		t.Logf("have: %s, want: %d, got: %d", have, want, got)
		t.Fail()
	}
}

func TestHeaderFields_ParseMD5(t *testing.T) {
	testHeader := newEmptyHeader()
	testFile, err := newLocalFile("filedefs/bytecode.cvd")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	testFileHeader, testFileBody := newSplit(testFile)
	testHeadParts := strings.Split(string(testFileHeader), ":")

	want := testHeadParts[5]

	testHeader.ParseMD5(testHeadParts[5], testFileBody)

	got := testHeader.MD5Hash

	t.Logf("got md5: %s, want: %s", got, want)

	if want != got && !testHeader.MD5Valid {
		t.Errorf("got md5: %s, want: %s", got, want)
		t.Fail()
	}
}
