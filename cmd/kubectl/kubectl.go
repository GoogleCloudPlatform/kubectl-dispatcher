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
	"strings"
	"syscall"

	"github.com/kubectl-dispatcher/pkg/client"
	"github.com/kubectl-dispatcher/pkg/filepath"
	"github.com/kubectl-dispatcher/pkg/util"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Timeout for server version query.
const requestTimeout = "5s"

// The kubectl dispatcher is a wrapper which retrieves the server version from
// a cluster, and executes the appropriate kubectl version. For example, if a
// user is configured to talk to their Kubernetes cluster that is version
// 1.10.3-gke, then this binary will execute "kubectl.1.10" (in the same
// directory as this binary) passing the arguments and environment of
// this binary.
//
// IMPORTANT: Versioned kubectl binaries that are dispatched to, MUST be in
// the "clibin" subdirectory to the current directory. Versioned kubectl binaries
// MUST follow the naming convention: kubectl.<major>.<minor>. Example:
// kubectl.1.12.
//
// NOTE: versioned kubectl filenames must NOT start with "kubectl-", since
// that is reserved for plugins. Therefore, we prefix versioned kubectl
// filenames with "kubectl.". Example: "kubectl.1.12"
func main() {
	// Create a defensive copy of the args and the environment.
	args := make([]string, len(os.Args))
	copy(args, os.Args)
	env := make([]string, len(os.Environ()))
	copy(env, os.Environ())

	// Initialize the flags: logs and kubeConfigFlags
	defer klog.Flush()
	usePersistentConfig := true
	kubeConfigFlags := genericclioptions.NewConfigFlags(usePersistentConfig)
	initFlags(kubeConfigFlags)

	// Create the default kubectl full file path.
	filepathBuilder := filepath.NewFilepathBuilder(&filepath.ExeDirGetter{}, os.Stat)
	kubectlDefaultFilepath := filepathBuilder.DefaultFilePath()
	kubectlFilepath := kubectlDefaultFilepath

	// Fetch the server version and generate the kubectl binary full file path
	// from this version.
	// Example:
	//   serverVersion=1.11 -> /home/seans/go/bin/kubectl.1.11
	svclient := client.NewServerVersionClient(kubeConfigFlags)
	svclient.SetRequestTimeout(requestTimeout)
	serverVersion, err := svclient.ServerVersion()
	if err == nil {
		klog.Infof("Server Version: %s", serverVersion.GitVersion)
		kubectlFilepath = filepathBuilder.VersionedFilePath(serverVersion)
		// Ensure this kubectl binary exists; otherwise fall back to default.
		if err := filepathBuilder.ValidateFilepath(kubectlFilepath); err != nil {
			klog.Warningf("Invalid kubectl filepath: %s (%v)", kubectlFilepath, err)
			// If default kubectl is also bad then fail. This should be the
			// only error the dispatcher surfaces.
			kubectlFilepath = kubectlDefaultFilepath
			if err := filepathBuilder.ValidateFilepath(kubectlFilepath); err != nil {
				klog.Errorf("Invalid default kubectl filepath: %s (%v) ", kubectlFilepath, err)
				os.Exit(1)
			}
		}
	} else {
		klog.Warningf("Error retrieving server version: (%v)", err)
		// If default kubectl is also bad then fail. This should be the
		// only error the dispatcher surfaces.
		if err := filepathBuilder.ValidateFilepath(kubectlFilepath); err != nil {
			klog.Errorf("Invalid default kubectl filepath: %s (%v) ", kubectlFilepath, err)
			os.Exit(1)
		}
	}

	// Delegate to the versioned or default kubectl binary. This overwrites the
	// current process (by calling execve(2) system call), and it does not return
	// on success.
	klog.Infof("kubectl dispatching: %s\n", kubectlFilepath)
	err = syscall.Exec(kubectlFilepath, args, env)
	if err != nil {
		klog.Errorf("kubectl dispatcher error: problem with Exec: (%v)", err)
	}
}

// Sets up the log flags and kubeConfigFlags. This dispatcher will pass most flags
// on to the versioned kubectl binary, so it must be resilient to unknown flags.
// The passed kubeConfigFlags must be connected to the pflagCommandList FlagSet,
// so the flag values can be filled in upon "Parse".
func initFlags(kubeConfigFlags *genericclioptions.ConfigFlags) {
	klog.InitFlags(flag.CommandLine)
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.CommandLine.SetNormalizeFunc(WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine) // Combine the flag and pflag FlagSets
	kubeConfigFlags.AddFlags(pflag.CommandLine)      // Binds kubeConfigFlags to the pflag FlagSet
	// Remove help flags, since these are special-cased in pflag.Parse,
	// and handled in the dispatcher instead of passed to versioned binary.
	args := os.Args[1:]
	args = util.RemoveAllElements(args, "-h")
	args = util.RemoveAllElements(args, "--help")
	pflag.CommandLine.Parse(args) // Fills in flags in FlagSet from args
	pflag.VisitAll(func(flag *pflag.Flag) {
		klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}

// WordSepNormalizeFunc changes all flags that contain "_" separators
// Copied from API Server
func WordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
	}
	return pflag.NormalizedName(name)
}
