#!/bin/bash
#############################################################
#
# NOTE: MUST be run from kubernetes root directory.
#
# Cross-compile kubectl into the necessary os/arch
# permutations. Creates an archive, and uploads these
# os/arch archives to Google Cloud Storage at the
# following location: gs://kubectl-dispatcher
#
#############################################################

set -e
set -x

VERSION="1.11.7"
DATE_TIME=$(date +%Y-%m-%d-%T)
SECONDS_EPOCH=$(date +'%s')

BASE_SOURCE_DIR="_output/dockerized/bin"
DEST_BUCKET="gs://kubectl-dispatcher"
DEST_DIR=${DEST_BUCKET}/v${VERSION}/${SECONDS_EPOCH}

echo "Building kubectl dispatcher for version: $VERSION"
echo "Date/Time: $DATE_TIME"
echo

# Clean up first
echo "Cleaning up: make clean"
echo
build/run.sh make clean
echo "Cleaning up: make test-cmd"
echo
build/run.sh make test-cmd
echo
echo

# Compile the linux variants
OS="linux"
ARCH="amd64"
echo "Building kubectl dispatcher: ${OS}/${ARCH}"
echo
build/run.sh make kubectl KUBE_BUILD_PLATFORMS=${OS}/${ARCH}
echo
echo
ARCHIVE_FILE="kubectl-dispatcher-${OS}-${ARCH}.tar.gz"
SOURCE_DIR="_output/dockerized/bin/${OS}/${ARCH}"
SOURCE_BIN="${SOURCE_DIR}/kubectl"
SOURCE_TAR="${SOURCE_DIR}/${ARCHIVE_FILE}"
DEST_TAR="${DEST_DIR}/${ARCHIVE_FILE}"
echo "Copying kubectl-dispatcher to Google Cloud Storage: $DEST_TAR"
tar cvzf $SOURCE_TAR $SOURCE_BIN
gsutil cp $SOURCE_TAR $DEST_TAR
echo
echo

ARCH="386"
echo "Building kubectl dispatcher: ${OS}/${ARCH}"
echo
build/run.sh make kubectl KUBE_BUILD_PLATFORMS=${OS}/${ARCH}
echo
echo
ARCHIVE_FILE="kubectl-dispatcher-${OS}-${ARCH}.tar.gz"
SOURCE_DIR="_output/dockerized/bin/${OS}/${ARCH}"
SOURCE_BIN="${SOURCE_DIR}/kubectl"
SOURCE_TAR="${SOURCE_DIR}/${ARCHIVE_FILE}"
DEST_TAR="${DEST_DIR}/${ARCHIVE_FILE}"
echo "Copying kubectl-dispatcher to Google Cloud Storage: $DEST_TAR"
tar cvzf $SOURCE_TAR $SOURCE_BIN
gsutil cp $SOURCE_TAR $DEST_TAR
echo
echo


# Compile the MacOS variants
OS="darwin"
ARCH="amd64"
echo "Building kubectl dispatcher: ${OS}/${ARCH}"
echo
build/run.sh make kubectl KUBE_BUILD_PLATFORMS=${OS}/${ARCH}
echo
echo
ARCHIVE_FILE="kubectl-dispatcher-${OS}-${ARCH}.tar.gz"
SOURCE_DIR="_output/dockerized/bin/${OS}/${ARCH}"
SOURCE_BIN="${SOURCE_DIR}/kubectl"
SOURCE_TAR="${SOURCE_DIR}/${ARCHIVE_FILE}"
DEST_TAR="${DEST_DIR}/${ARCHIVE_FILE}"
echo "Copying kubectl-dispatcher to Google Cloud Storage: $DEST_TAR"
tar cvzf $SOURCE_TAR $SOURCE_BIN
gsutil cp $SOURCE_TAR $DEST_TAR
echo
echo

ARCH="386"
echo "Building kubectl dispatcher: ${OS}/${ARCH}"
echo
build/run.sh make kubectl KUBE_BUILD_PLATFORMS=${OS}/${ARCH}
echo
echo
ARCHIVE_FILE="kubectl-dispatcher-${OS}-${ARCH}.tar.gz"
SOURCE_DIR="_output/dockerized/bin/${OS}/${ARCH}"
SOURCE_BIN="${SOURCE_DIR}/kubectl"
SOURCE_TAR="${SOURCE_DIR}/${ARCHIVE_FILE}"
DEST_TAR="${DEST_DIR}/${ARCHIVE_FILE}"
echo "Copying kubectl-dispatcher to Google Cloud Storage: $DEST_TAR"
tar cvzf $SOURCE_TAR $SOURCE_BIN
gsutil cp $SOURCE_TAR $DEST_TAR
echo
echo

# Compile the Windows variants
OS="windows"
ARCH="amd64"
echo "Building kubectl dispatcher: ${OS}/${ARCH}"
echo
build/run.sh make kubectl KUBE_BUILD_PLATFORMS=${OS}/${ARCH}
echo
echo
ARCHIVE_FILE="kubectl-dispatcher-${OS}-${ARCH}.tar.gz"
SOURCE_DIR="_output/dockerized/bin/${OS}/${ARCH}"
SOURCE_BIN="${SOURCE_DIR}/kubectl.exe"
SOURCE_TAR="${SOURCE_DIR}/${ARCHIVE_FILE}"
DEST_TAR="${DEST_DIR}/${ARCHIVE_FILE}"
echo "Copying kubectl-dispatcher to Google Cloud Storage: $DEST_TAR"
tar cvzf $SOURCE_TAR $SOURCE_BIN
gsutil cp $SOURCE_TAR $DEST_TAR
echo
echo


ARCH="386"
echo "Building kubectl dispatcher: ${OS}/${ARCH}"
echo
build/run.sh make kubectl KUBE_BUILD_PLATFORMS=${OS}/${ARCH}
echo
echo
ARCHIVE_FILE="kubectl-dispatcher-${OS}-${ARCH}.tar.gz"
SOURCE_DIR="_output/dockerized/bin/${OS}/${ARCH}"
SOURCE_BIN="${SOURCE_DIR}/kubectl.exe"
SOURCE_TAR="${SOURCE_DIR}/${ARCHIVE_FILE}"
DEST_TAR="${DEST_DIR}/${ARCHIVE_FILE}"
echo "Copying kubectl-dispatcher to Google Cloud Storage: $DEST_TAR"
tar cvzf $SOURCE_TAR $SOURCE_BIN
gsutil cp $SOURCE_TAR $DEST_TAR
echo
echo

