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
	"syscall"

	"github.com/kubectl-dispatcher/pkg/client"
	"github.com/kubectl-dispatcher/pkg/filepath"
	"github.com/kubectl-dispatcher/pkg/util"
	"github.com/spf13/pflag"
	utilflag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	// klog calls in this file assume it has been initialized beforehand
	"k8s.io/klog"
)

const requestTimeout = "5s" // Timeout for server version query
const cacheMaxAge = 60 * 60 // 1 hour in seconds

type Dispatcher struct {
	args            []string
	env             []string
	filepathBuilder *filepath.FilepathBuilder
}

// NewDispatcher returns a new pointer to a Dispatcher struct.
func NewDispatcher(args []string, env []string, filepathBuilder *filepath.FilepathBuilder) *Dispatcher {
	return &Dispatcher{
		args:            args,
		env:             env,
		filepathBuilder: filepathBuilder,
	}
}

// GetArgs returns a copy of the slice of strings representing the command line arguments.
func (d *Dispatcher) GetArgs() []string {
	return copyStrSlice(d.args)
}

// FilterArgs returns a copy of the slice of strings representing the command line arguments
// removing all the instances of each item in the passed "remove" slice. Used to return
// a slice of the command line args without flags in the "remove" slice.
func (d *Dispatcher) FilterArgs(remove []string) []string {
	args := d.GetArgs()
	for _, r := range remove {
		args = util.RemoveAllElements(args, r)
	}
	return args
}

// GetEnv returns a copy of the slice of environment variables.
func (d *Dispatcher) GetEnv() []string {
	return copyStrSlice(d.env)
}

func copyStrSlice(s []string) []string {
	c := make([]string, len(s))
	copy(c, s)
	return c
}

// InitKubeConfigFlags returns the ConfigFlags struct filled in with parsed
// kube config values parsed from command line arguments. These flag values can
// affect the server version query. Therefore, the set of kubeConfigFlags MUST
// match the set used in the regular kubectl binary.
func (d *Dispatcher) InitKubeConfigFlags() *genericclioptions.ConfigFlags {

	kubeConfigFlagSet := pflag.NewFlagSet("dispatcher", pflag.ExitOnError)
	kubeConfigFlagSet.ParseErrorsWhitelist.UnknownFlags = true
	kubeConfigFlagSet.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)

	usePersistentConfig := true
	kubeConfigFlags := genericclioptions.NewConfigFlags(usePersistentConfig)
	kubeConfigFlags.AddFlags(kubeConfigFlagSet)

	// Remove help flags, since these are special-cased in pflag.Parse,
	// and handled in the dispatcher instead of passed to versioned binary.
	args := d.FilterArgs([]string{"-h", "--help"})
	kubeConfigFlagSet.Parse(args[1:])
	kubeConfigFlagSet.VisitAll(func(flag *pflag.Flag) {
		klog.Infof("KubeConfig Flag: --%s=%q", flag.Name, flag.Value)
	})

	return kubeConfigFlags
}

// Dispatch attempts to execute a matching version of kubectl based on the
// version of the APIServer. If successful, this method will not return, since
// current process will be overwritten (see execve(2)). Otherwise, this method
// returns an error describing why the execution could not happen.
func (d *Dispatcher) Dispatch() error {
	// Fetch the server version and generate the kubectl binary full file path
	// from this version.
	// Example:
	//   serverVersion=1.11 -> /home/seans/go/bin/kubectl.1.11
	kubeConfigFlags := d.InitKubeConfigFlags()
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
	return syscall.Exec(kubectlFilepath, d.args, d.env)
}
