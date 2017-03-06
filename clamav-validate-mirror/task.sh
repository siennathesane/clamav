#!/bin/bash

set -eux
set -o pipefail

export ROOT=`pwd`
export RELEASE_PATH=`pwd`/clamav-release

# Download the clamav blob
pushd task-repo/concourse/tasks/clamav-validate-mirror
  bundle install
  bundle exec ruby download_blob.rb
  mv clamav-blob.tar.gz $ROOT
  mv pcre2-blob.tar.gz $ROOT
popd
# Create config file

cat << EOF > $ROOT/freshclam.conf
DatabaseDirectory $ROOT
UpdateLogFile $ROOT/freshclam.log
PidFile $ROOT/freshclam.pid
DatabaseOwner root
DatabaseMirror database.clamav.net
EOF

# Compile freshclam
tar xzf pcre2-blob.tar.gz
pushd pcre2-*
  ./configure --prefix=/tmp/pcre
  make
  make install
popd
tar xzf clamav-blob.tar.gz
pushd clamav-*.*.*
  ./configure --with-pcre=/tmp/pcre
  make
  ./freshclam/freshclam --config-file=$ROOT/freshclam.conf
popd
# Run the test against official mirror
