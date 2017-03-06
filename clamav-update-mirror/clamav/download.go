package clamav

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"net/http"
	"io/ioutil"
	"sync"
)

type VersionSet struct {
	Clamav       string
	Main         int
	Daily        int
	Safebrowsing int
	Bytecode     int
}

const Version = 1

const mainMirror = "http://database.clamav.net"
const clamavTxtRecord = "current.cvd.clamav.net"

var dbTypes = []string{"main", "bytecode", "daily"}

func ParseTxtRecordForVersions(versionString string) (VersionSet, error) {
	var clamavMajor, clamavMinor, clamavPatch, dummy int

	vs := VersionSet{}
	_, err := fmt.Sscanf(versionString, "%d.%d.%d:%d:%d:%d:%d:%d:%d:%d",
		&clamavMajor,
		&clamavMinor,
		&clamavPatch,
		&vs.Main,
		&vs.Daily,
		&dummy, &dummy, &dummy,
		&vs.Safebrowsing,
		&vs.Bytecode)

	if err != nil {
		return vs, err
	}

	vs.Clamav = fmt.Sprintf("%d.%d.%d", clamavMajor, clamavMinor, clamavPatch)

	return vs, nil
}

func ParseCvdVersion(cvdFile string) (int, error) {
	fh, err := os.Open(cvdFile)
	if err != nil {
		return 0, nil
	}
	defer fh.Close()

	head := make([]byte, 512)
	_, err = fh.Read(head)
	if err != nil {
		return 0, nil
	}
	headStr := string(head)
	headParts := strings.Split(headStr, ":")
	if len(headParts) < 3 {
		return 0, fmt.Errorf("Invalid header string: %s", headStr)
	}

	var verNum int
	verNum, err = strconv.Atoi(headParts[2])
	if err != nil {
		return 0, err
	}

	return verNum, nil
}

func DownloadDatabase(dbDir string) error {
	var err error
	var downloadClient  = &http.Client{}
	var wg sync.WaitGroup

	for _, dbType := range dbTypes {
		wg.Add(1)
		go downloadFile(mainMirror+"/"+dbType+".cvd", dbDir, downloadClient, &wg)

		var cdiffVer int
		cdiffVer, err = ParseCvdVersion(filepath.Join(dbDir, dbType+".cvd"))
		if err != nil {
			return err
		}
		cdiffUrl := mainMirror + "/" + dbType + "-" + strconv.Itoa(cdiffVer) + ".cdiff"
		wg.Add(1)
		go downloadFile(cdiffUrl, dbDir, downloadClient, &wg)
	}
	wg.Wait()

	return nil
}

func downloadFile(rawUrl, dest string, cl *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	var err error
	var cvdUrl *url.URL
	if cvdUrl, err = url.Parse(rawUrl); err != nil {
		panic(err)
	}

	filename := strings.TrimLeft(cvdUrl.Path, "/")
	log.Println("Downloading", filename)

	cvdDest := filepath.Join(dest, filename)
	resp, err := cl.Get(rawUrl)
	if err != nil {
		log.Panicf("failed to download file! %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicf("failed to read file! %s", err)
	}

	if err = ioutil.WriteFile(cvdDest, body, 0666); err != nil {
		log.Panicf("failed to write file! %s", err)
	}
	log.Println("done downloading", cvdDest)
}
