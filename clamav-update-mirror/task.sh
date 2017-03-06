#!/bin/bash

set -eux
set -o pipefail

export DB_DIR=$PWD/cvds

mkdir -p $DB_DIR

date +%s > $DB_DIR/last

export GOPATH="${PWD}/go"
pushd "${GOPATH}/src/github.com/pivotal-cloudops/cloudops-ci/concourse/tasks/clamav-update-mirror" > /dev/null
  go get -v ./...
  go run main.go $DB_DIR
popd > /dev/null
