#!/bin/bash

export ALLOW_SKEW=1

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
#git checkout .
#git checkout master
#git branch -D branch-v${BRANCH}
echo
echo "FINISHED"
