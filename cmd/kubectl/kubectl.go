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
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	// Create a defensive copy of the args and the environment.
	args := make([]string, len(os.Args))
	copy(args, os.Args)
	env := make([]string, len(os.Environ()))
	copy(env, os.Environ())

	// Initialize the flags: logs and kubeConfigFlags
	defer klog.Flush()
	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	initFlags(kubeConfigFlags)

	// Using the kube config flags values, create the discovery client and contact
	// the api server to retrieve the version.
	klog.Info("Creating discovery client")
	discoveryClient, err := kubeConfigFlags.ToDiscoveryClient()
	if err != nil {
		klog.Errorf("kubectl dispatcher error: unable to create discovery client (%v)", err)
		os.Exit(1)
	}
	serverVersionClient, err := client.NewServerVersionClient(discoveryClient)
	if err != nil {
		klog.Errorf("kubectl dispatcher error: error creating server version client (%v)", err)
		os.Exit(1)
	}
	serverVersion, err := serverVersionClient.ServerVersion()
	if err != nil {
		klog.Errorf("kubectl dispatcher error: error retrieving server version (%v)", err)
		os.Exit(1)
	}
	klog.Infof("Server Version: %s", serverVersion.GitVersion)

	// Create the full versioned kubectl file path from the server version, and
	// the current directory of this dispatcher binary.
	// Example:
	//   serverVersion -> /home/seans/go/bin/kubectl-1.11
	klog.Info("Creating versioned kubectl binary full file path")
	filepathBuilder, err := filepath.NewFilepathBuilder(serverVersion, &filepath.ExeDirGetter{})
	if err != nil {
		klog.Errorf("kubectl dispatcher error: error creating filepath builder (%v)", err)
	}
	kubectlFilepath, err := filepathBuilder.VersionedFilePath()
	if err != nil {
		klog.Errorf("kubectl dispatcher error: error creating kubectl binary file path: (%v)", err)
		os.Exit(1)
	}
	if _, err := os.Stat(kubectlFilepath); err != nil {
		klog.Errorf("kubectl dispatcher error: unable to find kubectl executable (%s)", kubectlFilepath)
		os.Exit(1)
	}

	// Dispatch to the versioned kubectl binary.
	klog.Infof("kubectl dispatching: %s\n", kubectlFilepath)
	execErr := syscall.Exec(kubectlFilepath, args, env)
	if execErr != nil {
		klog.Errorf("kubectl dispatcher error: error executing kubectl binary (%v)", execErr)
		os.Exit(1)
	}
	klog.Info("kubectl dispatcher complete")
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
	pflag.CommandLine.Parse(os.Args[1:])             // Fills in flags in FlagSet from args
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
