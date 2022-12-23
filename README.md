## ClamAV Private Mirror

[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![Build Status](https://travis-ci.org/mxplusb/clamav.svg?branch=master)](https://travis-ci.org/mxplusb/clamav)

### How To Use

* `cf push`

OR

* `go get -v ./... && go build -v . && PRIMARY_MIRROR="https://database.clamav.net" ./clamav`

### What It Does

1. Starts an asynchronous download of the current antivirus definitions.
    1. Downloads three databases:
        1. `main`
        1. `bytecode`
        1. `daily`
    1. Parses each database's header for similar versions.
    1. If there is a similar/related version, it also gets downloaded.
    1. Downloaded files are stored in-memory in a cache for client downloads.
1. Initialises a cron job to download the new database definitions every hour.
1. Starts the web server and serves from cache.
1. Evicts files from cache every 3 hours to prevent stale definitions.

### Mirrors

In order to function properly as a localised cache, you need to set the `PRIMARY_MIRROR` environment variable. Below is a short list of known mirrors.

* http://database.clamav.net

If for some reason the primary mirror fails, if you set `SECONDARY_MIRROR`, it will try that one.

### Contributing

* Keep It Simple.

To unzip the CVD files:
`cd filedefs/ && tail -c $(expr $(wc -c $FILE.cvd | awk '{print $1}') - 512) $FILE.cvd | tar zxvf -`