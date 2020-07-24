#!/bin/bash
##################################################################
# Copyright 2019 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Cross-compile kubectl into the necessary os/arch
# permutations. Creates an archive, and uploads these
# os/arch archives to Google Cloud Storage at the
# following location: gs://kubectl-dispatcher
#
# NOTE: MUST be run from kubernetes root directory.
#
##################################################################

set -e
set -x

VERSION="1.15.12"
DISPATCHER_VERSION="1.0"
DATE_TIME=$(date +%Y-%m-%d-%T)
SECONDS_EPOCH=$(date +'%s')

OSES=("linux" "darwin" "windows")
ARCHES=("amd64" "386")

WINDOWS_SUFFIX=".exe"

BASE_SOURCE_DIR="_output/dockerized/bin"
DEST_BUCKET="gs://kubectl-dispatcher"
DEST_DIR=${DEST_BUCKET}/v${VERSION}/${SECONDS_EPOCH}
DEST_RELEASE=${DEST_BUCKET}/v${VERSION}/release

echo "Building kubectl dispatcher for version: $VERSION"
echo "Date/Time: $DATE_TIME"
echo

# Check if the git status is clean.
if [ -z "$(git status --porcelain)" ]; then 
  echo "git status clean--continue"
else 
  echo "Uncommitted changes in current directory--exiting"
  exit 1
fi

# Tag the build
echo "Tag the kubectl dispatcher build"
set +e
git tag -d "v${VERSION}-dispatcher" 2>&1
set -e
git tag -a "v${VERSION}-dispatcher" -m "kubectl dispatcher v${DISPATCHER_VERSION} at fork of v${VERSION}"
echo
echo

# Clean up first
echo "Cleaning up: make clean"
echo
build/run.sh make clean
echo "Cleaning up: make test-cmd"
echo
# build/run.sh make test-cmd
echo
echo

# For each os/arch combination, we build the kubectl binary, then
# package it and upload to our GCS bucket.
for OS in ${OSES[*]}
do
  for ARCH in ${ARCHES[*]}
  do
    SOURCE_DIR="_output/dockerized/bin/${OS}/${ARCH}"
    SOURCE_BIN="${SOURCE_DIR}/kubectl-sdk"
    DEST_BIN="${SOURCE_DIR}/kubectl"
    # In windows, kubectl binary is named "kubectl.exe"
    if [ $OS = "windows" ]; then
      SOURCE_BIN="${SOURCE_BIN}${WINDOWS_SUFFIX}"
      DEST_BIN="${DEST_BIN}${WINDOWS_SUFFIX}"
    fi
    echo "Building static kubectl dispatcher: ${OS}/${ARCH}"
    echo
    build/run.sh make kubectl-sdk KUBE_BUILD_PLATFORMS=${OS}/${ARCH} CGO_ENABLED=0
    mv ${SOURCE_BIN} ${DEST_BIN}
    echo
    echo
    ARCHIVE_FILE="kubectl-dispatcher-${OS}-${ARCH}.tar.gz"
    SOURCE_TAR="${SOURCE_DIR}/${ARCHIVE_FILE}"
    DEST_TAR="${DEST_DIR}/${ARCHIVE_FILE}"
    RELEASE_TAR="${DEST_RELEASE}/${ARCHIVE_FILE}"
    echo "Copying kubectl-dispatcher to Google Cloud Storage: $DEST_TAR"
    tar cvzf $SOURCE_TAR $DEST_BIN
    gsutil cp $SOURCE_TAR $DEST_TAR
    gsutil cp $SOURCE_TAR $RELEASE_TAR
    echo
    echo
  done
done
