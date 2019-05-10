/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"
	"syscall"

	"github.com/GoogleCloudPlatform/kubectl-dispatcher/pkg/dispatcher"
	"github.com/GoogleCloudPlatform/kubectl-dispatcher/pkg/filepath"
	"github.com/GoogleCloudPlatform/kubectl-dispatcher/pkg/util"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/version"
	utilflag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/klog"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Hard-coded default client version.
var clientVersion = version.Info{
	Major:      "1",
	Minor:      "11",
	GitVersion: "v1.11.7",
}

// The kubectl dispatcher is a wrapper which retrieves the server version from
// a cluster, and executes the appropriate kubectl version. For example, if a
// user is configured to talk to their Kubernetes cluster that is version
// 1.10.3-gke, then this binary will execute "kubectl.1.10" (in the same
// directory as this binary) passing the arguments and environment of
// this binary.
//
// IMPORTANT: Versioned kubectl binaries that are dispatched to, MUST be in
// the the current directory. Versioned kubectl binaries MUST follow the
// naming convention: kubectl.<major>.<minor>. Example: kubectl.1.12.
//
// NOTE: versioned kubectl filenames must NOT start with "kubectl-", since
// that is reserved for plugins. Therefore, we prefix versioned kubectl
// filenames with "kubectl.". Example: "kubectl.1.12"
func main() {

	InitLogging(os.Args[1:])
	defer klog.Flush()

	// Dispatch() does not return if successful; the current process is overwritten.
	dispatcher.Execute(clientVersion)

	// Dispatch to the default kubectl binary given by clientVersion.
	filepathBuilder := filepath.NewFilepathBuilder(&filepath.ExeDirGetter{}, os.Stat)
	kubectlDefaultFilepath, err := filepathBuilder.VersionedFilePath(clientVersion)
	if err != nil {
		klog.Errorf("Error creating default kubectl: (%v)", err)
		os.Exit(1)
	}
	if err := filepathBuilder.ValidateFilepath(kubectlDefaultFilepath); err != nil {
		klog.Errorf("Error validating default kubectl: %s (%v)", kubectlDefaultFilepath, err)
		os.Exit(1)
	}

	klog.Infof("Default kubectl dispatched: %s", kubectlDefaultFilepath)
	err = syscall.Exec(kubectlDefaultFilepath, os.Args, os.Environ())
	if err != nil {
		klog.Errorf("kubectl dispatcher error: problem with Exec: (%v)", err)
	}
}

// Initialize klog logging by parsing the log-related flags.
func InitLogging(args []string) {
	logFlagSet := flag.NewFlagSet("dispatcher-logs", flag.ContinueOnError)
	klog.InitFlags(logFlagSet)
	plogFlagSet := pflag.NewFlagSet("dispatcher-kube-config", pflag.ContinueOnError)
	plogFlagSet.ParseErrorsWhitelist.UnknownFlags = true
	plogFlagSet.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	plogFlagSet.AddGoFlagSet(logFlagSet)
	// Remove help flags, pflags handles them specially.
	args = util.FilterList(args, dispatcher.HelpFlags)
	plogFlagSet.Parse(args)
}
