#!/bin/bash
#############################################################
#
# Run end-to-end tests for kubectl with dispatcher. Runs
# against several versions of the control plane, and
# outputs results to file in temp directory.
#
#############################################################

set -e

# Validate that the kubectl binary exists
KUBECTL_BIN=${HOME}/go/src/k8s.io/kubernetes/_output/local/bin/linux/amd64/kubectl
if [ ! -f $KUBECTL_BIN ]; then
    echo "ERROR: kubectl dispatch binary not found--exiting"
    exit 1
fi

BRANCH_1_9="1.9"
BRANCH_1_10="1.10"
BRANCH_1_11="1.11"
BRANCH_1_12="1.12"
BRANCH_1_13="1.13"

BRANCHES=($BRANCH_1_9 $BRANCH_1_10 $BRANCH_1_11 $BRANCH_1_12 $BRANCH_1_13)

# Create the temporary directory for the binaries and logs of our test.
DATE_TIME=$(date +%Y-%m-%d-%T)
CURRENT_DIR=$(pwd)
TMPDIR=$(mktemp -d --tmpdir=$CURRENT_DIR "$DATE_TIME-XXXXXXXX")
BIN_DIR=${TMPDIR}/bin
mkdir $BIN_DIR
LOG_DIR=${TMPDIR}/logs
mkdir $LOG_DIR

# Copy the kubectl dispatcher into the BIN_DIR
echo "Copy $KUBECTL_BIN to $BIN_DIR"
echo
cp $KUBECTL_BIN $BIN_DIR
KUBECTL_TEST=$BIN_DIR/kubectl

# Download all the versioned kubectl binaries
VERSION="v1.9.11"
SHORT_VERSION="1.9"
echo "Version: ${VERSION}"
wget https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl ${BIN_DIR}/kubectl.${SHORT_VERSION}

VERSION="v1.10.12"
SHORT_VERSION="1.10"
echo "Version: ${VERSION}"
wget https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl ${BIN_DIR}/kubectl.${SHORT_VERSION}

VERSION="v1.11.7"
SHORT_VERSION="1.11"
echo "Version: ${VERSION}"
wget https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl ${BIN_DIR}/kubectl.${SHORT_VERSION}

VERSION="v1.12.5"
SHORT_VERSION="1.12"
echo "Version: ${VERSION}"
wget https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl ${BIN_DIR}/kubectl.${SHORT_VERSION}

VERSION="v1.13.3"
SHORT_VERSION="1.13"
echo "Version: ${VERSION}"
wget https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl ${BIN_DIR}/kubectl.${SHORT_VERSION}


echo "${DATE_TIME}: End-to-End Tests for kubectl dispatcher"
echo
for BRANCH in ${BRANCHES[*]}
do
  outfile="$LOG_DIR/$BRANCH-end-to-end.out"
  exec &> $outfile
  DATE_TIME=$(date +%Y-%m-%d-%T)
  echo "${DATE_TIME}: Branch: $BRANCH"
  echo
  DATE_TIME=$(date +%Y-%m-%d-%T)
  echo "${DATE_TIME}: kubetest --extract=release/stable-${BRANCH} --up --test --down --test_args=\"--kubectl-path=${KUBECTL_TEST} --ginkgo.focus=\[sig\-cli\]\""
  # kubetest --extract=release/stable-${BRANCH} --up --test --down --test_args="--kubectl-path=${KUBECTL_BIN} --ginkgo.focus=\[sig\-cli\]"
done

echo
echo
DATE_TIME=$(date +%Y-%m-%d-%T)
echo "${DATE_TIME}: Finished"
