## ClamAV Private Mirror

[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)

### How To Use

* `cf push`

OR

* `glide install && go build -v . && ./clamav`

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

### TODO

1. Finish header structs (including tests).
1. Parse file on download, before cache insertion.
1. Don't return bad files.

### Contributing

* Keep It Simple.
