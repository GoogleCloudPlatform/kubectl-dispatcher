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

package filepath

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubectl-dispatcher/pkg/util"
	"k8s.io/apimachinery/pkg/version"

	// klog calls in this file assume it has been initialized beforehand
	"k8s.io/klog"
)

// DirectoryGetter implements a single function returning the "current directory".
type DirectoryGetter interface {
	CurrentDirectory() (string, error)
}

// ExeDirGetter implements the DirectoryGetter interface.
type ExeDirGetter struct{}

// CurrentDirectory returns the absolute full directory path of the
// currently running executable.
func (e *ExeDirGetter) CurrentDirectory() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	abs, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}
	return filepath.Dir(abs), nil
}

// FilepathBuilder encapsulates the data and functionality to build the full
// versioned kubectl filepath from the server version.
type FilepathBuilder struct {
	dirGetter DirectoryGetter
	// Function to call to check if a file exists.
	filestatFunc func(string) (os.FileInfo, error)
}

// NewFilepathBuilder encapsulates information necessary to build the full
// file path of the versioned kubectl binary to execute. NOTE: A nil
// ServerVersion is acceptable, and it maps to the default kubectl version.
func NewFilepathBuilder(dirGetter DirectoryGetter, filestat func(string) (os.FileInfo, error)) *FilepathBuilder {
	return &FilepathBuilder{
		dirGetter:    dirGetter,
		filestatFunc: filestat,
	}
}

// currentDir returns the full file path to the directory for storing the
// versioned kubectl binaries (e.g. kubectl.1.12, kubectl.default).
func (c *FilepathBuilder) currentDir() string {
	currentDirectory := "" // Use empty directory upon error.
	if c.dirGetter != nil {
		dir, err := c.dirGetter.CurrentDirectory()
		if err != nil {
			klog.Warningf("kubectl dispatcher current directory error: (%v)", err)
		} else {
			currentDirectory = dir
		}
	} else {
		klog.Warningf("directory getter is nil; using empty current directory")
	}
	return currentDirectory
}

// VersionedFilePath returns the full absolute file path to the
// versioned kubectl binary to dispatch to. On error, empty string is returned.
func (c *FilepathBuilder) VersionedFilePath(serverVersion *version.Info) string {
	kubectlFilename := ""
	if serverVersion != nil {
		majorVersion, err := util.GetMajorVersion(serverVersion)
		if err == nil {
			minorVersion, err := util.GetMinorVersion(serverVersion)
			if err == nil {
				// Example: major: "1", minor: "12" -> "kubectl.1.12"
				kubectlFilename, err = createKubectlBinaryFilename(majorVersion, minorVersion)
			} else {
				klog.Warningf("Error generating minor version number: (%v)", err)
			}
		} else {
			klog.Warningf("Error generating major version number: (%v)", err)
		}
	} else {
		klog.Warningf("Server version is nil while generating versioned file path")
	}
	return filepath.Join(c.currentDir(), kubectlFilename)
}

func (c *FilepathBuilder) ValidateFilepath(filepath string) error {
	if _, err := c.filestatFunc(filepath); err != nil {
		return err
	}
	return nil
}

const kubectlBinaryName = "kubectl"

// NOTE: versioned kubectl filenames must NOT start with "kubectl-", since
// that is reserved for plugins. Therefore, we prefix versioned kubectl
// filenames with "kubectl.". Example: "kubectl.1.12"
func createKubectlBinaryFilename(major string, minor string) (string, error) {
	return fmt.Sprintf("%s.%s.%s", kubectlBinaryName, major, minor), nil
}
