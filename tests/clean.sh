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
##################################################################

echo "Cleaning; make clean"
make clean
echo

# NOTE: You may have to recreate dir: /var/run/kubernetes

DISPATCHER_BIN="${HOME}/go/src/github.com/kubectl-dispatcher"
KUBECTL_BIN="_output/local/bin/linux/amd64"
# OUTPUT_BIN="${KUBECTL_BIN}/clibin"
OUTPUT_BIN="${KUBECTL_BIN}"

echo "Copying kubectl dispatcher to: ${OUTPUT_BIN}/kubectl.dispatcher"
mkdir -p ${OUTPUT_BIN}
cp -f ${DISPATCHER_BIN}/kubectl ${OUTPUT_BIN}/kubectl.dispatcher
echo "Copying kubectl dispatcher to: ${KUBECTL_BIN}/kubectl"
cp -f ${OUTPUT_BIN}/kubectl.dispatcher ${KUBECTL_BIN}/kubectl

echo "Downloading versioned kubectl binaries"
echo

VERSION="v1.9.11"
SHORT_VERSION="1.9"
echo "Version: ${VERSION}"
wget https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl ${OUTPUT_BIN}/kubectl.${SHORT_VERSION}

VERSION="v1.10.7"
SHORT_VERSION="1.10"
echo "Version: ${VERSION}"
wget https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl ${OUTPUT_BIN}/kubectl.${SHORT_VERSION}

VERSION="v1.11.7"
SHORT_VERSION="1.11"
echo "Version: ${VERSION}"
wget https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl ${OUTPUT_BIN}/kubectl.${SHORT_VERSION}

VERSION="v1.12.5"
SHORT_VERSION="1.12"
echo "Version: ${VERSION}"
wget https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl ${OUTPUT_BIN}/kubectl.${SHORT_VERSION}

VERSION="v1.13.3"
SHORT_VERSION="1.13"
echo "Version: ${VERSION}"
wget https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
mv kubectl ${OUTPUT_BIN}/kubectl.${SHORT_VERSION}

VERSION="v1.14.0"
SHORT_VERSION="1.14"
echo "Version: ${VERSION}"
cp bazel-bin/cmd/kubectl/linux_amd64_pure_stripped/kubectl ${OUTPUT_BIN}/kubectl.${SHORT_VERSION}
echo

DEFAULT_VERSION="1.11"
echo 
echo "Version ${DEFAULT_VERSION} is default"
echo "Creating default kubectl version"
cp ${OUTPUT_BIN}/kubectl.${DEFAULT_VERSION} ${OUTPUT_BIN}/kubectl.default
echo

