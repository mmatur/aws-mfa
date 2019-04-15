#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -e

APP_DIR="/go/src/github.com/${GITHUB_REPOSITORY}/"

export GO111MODULE=on

mkdir -p ${APP_DIR} && cp -r ./ ${APP_DIR} && cd ${APP_DIR}


make $@