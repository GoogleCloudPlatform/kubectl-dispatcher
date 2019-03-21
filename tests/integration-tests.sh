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
# Runs integration tests (make test-cmd) for kubectl
# dispatcher for various branches.
#
# NOTE: MUST be run from the kubernete root dir (e.g.
#   ~/go/src/k8s.io/kubernetes).
#
##################################################################

set -e

export ALLOW_SKEW=1

# Validate that the kubectl binary exists
KUBECTL_BIN=${HOME}/go/src/k8s.io/kubernetes/_output/dockerized/bin/linux/amd64/kubectl
if [ ! -f $KUBECTL_BIN ]; then
    echo "ERROR: kubectl dispatch binary not found--exiting"
    exit 1
fi

# BRANCH="1.9.11"
# BRANCH="1.10.12"
# BRANCH="1.11.7"
BRANCH="1.12.5"
# BRANCH="1.13.3"

OUTPUT_DIR="_output/local/bin/linux/amd64"
CLIBIN_DIR="${OUTPUT_DIR}"

echo "========================="
echo "Checking out branch ${BRANCH}"
echo "========================="
echo
git checkout v${BRANCH} -b branch-v${BRANCH}
echo
echo

echo "========================="
echo "Cleaning and Setup"
echo "========================="
echo
./clean.sh
echo 
echo

echo "========================="
echo "Setting kubectl dispatcher"
echo "========================="
echo
sleep 5
# CMD="cp ${CLIBIN_DIR}/kubectl.dispatcher ${OUTPUT_DIR}/kubectl\n"
CMD="cp ${CLIBIN_DIR}/kubectl.1.10 ${OUTPUT_DIR}/kubectl\n"
MODIFY_FILE="hack/make-rules/test-cmd.sh"
sed -i -e "s|^runTests|${CMD}runTests|g" ${MODIFY_FILE}
echo 
echo

echo "========================="
echo "Starting integration test"
echo "========================="
echo
make test-cmd
echo 
echo

echo "========================="
echo "Removing branch ${BRANCH}"
echo "========================="
echo
git checkout .
git checkout master
git branch -D branch-v${BRANCH}
echo
echo "FINISHED"
