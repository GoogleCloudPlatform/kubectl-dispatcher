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
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// The kubectl dispatcher is a wrapper which retrieves the server version from
// a cluster, and executes the appropriate kubectl version. For example, if a
// user is configured to talk to their Kubernetes cluster that is version
// 1.10.3-gke, then this binary will execute "kubectl.1.10" (in the same
// directory as this binary) passing the arguments and environment of
// this binary.
//
// IMPORTANT: Versioned kubectl binaries that are dispatched to, MUST be in
// the same directory as this dispatcher binary. Versioned kubectl binaries
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

	filepathBuilder := filepath.NewFilepathBuilder(&filepath.ExeDirGetter{})
	kubectlDefaultFilepath := filepathBuilder.DefaultFilePath()

	// Fetch the server version; nil implies using the default version of kubectl.
	// Then create the full versioned kubectl file path from the server version, and
	// the current directory of this dispatcher binary.
	// Example:
	//   serverVersion=1.11 -> /home/seans/go/bin/kubectl.1.11
	kubectlFilepath := kubectlDefaultFilepath
	if serverVersion := getServerVersion(kubeConfigFlags); serverVersion != nil {
		klog.Infof("Server Version: %s", serverVersion.GitVersion)
		kubectlFilepath = filepathBuilder.VersionedFilePath(serverVersion)
	}
	// Fall back to default kubectl binary if versioned kubectl is bad.
	if !validateFilepath(kubectlFilepath) {
		klog.Infof("Invalid kubectl filepath: %s", kubectlFilepath)
		kubectlFilepath = kubectlDefaultFilepath
		// If default kubectl is also bad then fail. This is should be the
		// only error the dispatcher surfaces.
		if !validateFilepath(kubectlFilepath) {
			klog.Errorf("Invalid default kubectl filepath: %s", kubectlFilepath)
			os.Exit(1)
		}
	}

	// Delegate to the versioned or default kubectl binary. This overwrites the
	// current process (by calling execve(2) system call), and it does not return
	// on success.
	klog.Infof("kubectl dispatching: %s\n", kubectlFilepath)
	err := syscall.Exec(kubectlFilepath, args, env)
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

// getServerVersion returns the server version of the Kubernetes cluster, or
// nil if there is an error.
func getServerVersion(kubeConfigFlags *genericclioptions.ConfigFlags) *version.Info {
	// Using the kube config flags values, create the discovery client and contact
	// the api server to retrieve the version.
	klog.Info("Creating discovery client")
	discoveryClient, err := kubeConfigFlags.ToDiscoveryClient()
	if err != nil {
		klog.Infof("kubectl dispatcher error: unable to create discovery client (%v)", err)
		return nil
	}
	serverVersionClient, err := client.NewServerVersionClient(discoveryClient)
	if err != nil {
		klog.Infof("kubectl dispatcher error: error creating server version client (%v)", err)
		return nil
	}
	serverVersion, err := serverVersionClient.ServerVersion()
	if err != nil {
		klog.Infof("kubectl dispatcher error: error retrieving server version (%v)", err)
		return nil
	}
	return serverVersion
}

func validateFilepath(filepath string) bool {
	isValid := true
	if _, err := os.Stat(kubectlFilepath); err != nil {
		isValid = false
	}
	return valid
}
