#!/bin/bash

set -eux

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

CVDS=$(mktemp -d -t stemcell)

fly \
  -t ci \
  execute \
  -c "${SCRIPT_DIR}/task.yml" \
  -i task-repo="${SCRIPT_DIR}/../.." \
  -o cvds="${CVDS}"
