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

package dispatcher

import (
	"flag"
	"strings"
	"syscall"

	"github.com/kubectl-dispatcher/pkg/client"
	"github.com/kubectl-dispatcher/pkg/filepath"
	"github.com/kubectl-dispatcher/pkg/util"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
)

const requestTimeout = "5s" // Timeout for server version query
const cacheMaxAge = 60 * 60 // 1 hour in seconds

type Dispatcher struct {
	args            []string
	env             []string
	filepathBuilder *filepath.FilepathBuilder
}

func NewDispatcher(args []string, env []string, filepathBuilder *filepath.FilepathBuilder) *Dispatcher {
	return &Dispatcher{
		args:            copyStrSlice(args),
		env:             copyStrSlice(env),
		filepathBuilder: filepathBuilder,
	}
}

func copyStrSlice(s []string) []string {
	c := make([]string, len(s))
	copy(c, s)
	return c
}

func (d *Dispatcher) InitFlags() *genericclioptions.ConfigFlags {

	usePersistentConfig := true
	kubeConfigFlags := genericclioptions.NewConfigFlags(usePersistentConfig)

	klog.InitFlags(flag.CommandLine)
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.CommandLine.SetNormalizeFunc(wordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine) // Combine the flag and pflag FlagSets
	kubeConfigFlags.AddFlags(pflag.CommandLine)      // Binds kubeConfigFlags to the pflag FlagSet
	// Remove help flags, since these are special-cased in pflag.Parse,
	// and handled in the dispatcher instead of passed to versioned binary.
	argsCopy := copyStrSlice(d.args[1:])
	argsCopy = util.RemoveAllElements(argsCopy, "-h")
	argsCopy = util.RemoveAllElements(argsCopy, "--help")
	pflag.CommandLine.Parse(argsCopy) // Fills in flags in FlagSet from args
	pflag.VisitAll(func(flag *pflag.Flag) {
		klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})

	return kubeConfigFlags
}

// WordSepNormalizeFunc changes all flags that contain "_" separators
// Copied from API Server
func wordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
	}
	return pflag.NormalizedName(name)
}

func (d *Dispatcher) Dispatch() error {
	// Initialize the flags: logs and kubeConfigFlags
	defer klog.Flush()
	kubeConfigFlags := d.InitFlags()

	// Fetch the server version and generate the kubectl binary full file path
	// from this version.
	// Example:
	//   serverVersion=1.11 -> /home/seans/go/bin/kubectl.1.11
	svclient := client.NewServerVersionClient(kubeConfigFlags)
	svclient.SetRequestTimeout(requestTimeout)
	svclient.SetCacheMaxAge(cacheMaxAge)
	serverVersion, err := svclient.ServerVersion()
	if err != nil {
		return err
	}
	klog.Infof("Server Version: %s", serverVersion.GitVersion)
	kubectlFilepath := d.filepathBuilder.VersionedFilePath(serverVersion)
	// Ensure the versioned kubectl binary exists.
	if err := d.filepathBuilder.ValidateFilepath(kubectlFilepath); err != nil {
		return err
	}

	// Delegate to the versioned or default kubectl binary. This overwrites the
	// current process (by calling execve(2) system call), and it does not return
	// on success.
	klog.Infof("kubectl dispatching: %s\n", kubectlFilepath)
	err = syscall.Exec(kubectlFilepath, d.args, d.env)
	if err != nil {
		return err
	}
	return nil
}
