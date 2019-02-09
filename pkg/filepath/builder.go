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
	"strconv"
	"strings"
	"unicode"

	"k8s.io/apimachinery/pkg/version"
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
}

// NewFilepathBuilder encapsulates information necessary to build the full
// file path of the versioned kubectl binary to execute. NOTE: A nil
// ServerVersion is acceptable, and it maps to the default kubectl version.
func NewFilepathBuilder(dirGetter DirectoryGetter) *FilepathBuilder {
	return &FilepathBuilder{
		dirGetter: dirGetter,
	}
}

const kubectlDefaultName = "kubectl.default"

func (c *FilepathBuilder) currentDirectory() string {
	currentDirectory := "" // Use empty directory upon error.
	if c.dirGetter != nil {
		dir, err := c.dirGetter.CurrentDirectory()
		if err != nil {
			klog.Infof("kubectl dispatcher current directory error: (%v)", err)
		} else {
			currentDirectory = dir
		}
	} else {
		klog.Infof("directory getter is nil; using empty current directory")
	}
	return currentDirectory
}

func (c *FilepathBuilder) DefaultFilePath() string {
	return filepath.Join(c.currentDirectory(), kubectlDefaultName)
}

// VersionedFilePath returns the full absolute file path to the
// versioned kubectl binary to dispatch to. On error, the full path to the
// default kubectl binary is returned.
func (c *FilepathBuilder) VersionedFilePath(serverVersion *version.Info) string {
	// Use default filename upon error.
	kubectlFilename := kubectlDefaultName
	if serverVersion != nil {
		majorVersion, err := getMajorVersion(serverVersion)
		if err == nil {
			minorVersion, err := getMinorVersion(serverVersion)
			if err == nil {
				// Example: major: "1", minor: "12" -> "kubectl.1.12"
				kubectlFilename, err = createKubectlBinaryFilename(majorVersion, minorVersion)
			} else {
				klog.Infof("kubectl dispatching default binary: (%v)", err)
			}
		} else {
			klog.Infof("kubectl dispatching default binary: (%v)", err)
		}
	} else {
		klog.Infof("kubectl dispatching default binary: server version is nil")
	}
	return filepath.Join(c.currentDirectory(), kubectlFilename)
}

func getMajorVersion(serverVersion *version.Info) (string, error) {
	if serverVersion == nil {
		return "", fmt.Errorf("server version is nil")
	}
	majorStr, err := normalizeVersionStr(serverVersion.Major)
	if err != nil {
		return "", err
	}
	if !isPositiveInteger(majorStr) {
		return "", fmt.Errorf("Bad major version string: %s", majorStr)
	}
	return majorStr, nil
}

func getMinorVersion(serverVersion *version.Info) (string, error) {
	if serverVersion == nil {
		return "", fmt.Errorf("server version is nil")
	}
	minorStr, err := normalizeVersionStr(serverVersion.Minor)
	if err != nil {
		return "", err
	}
	if !isPositiveInteger(minorStr) {
		return "", fmt.Errorf("Bad minor version string: %s", minorStr)
	}
	return minorStr, nil
}

const kubectlBinaryName = "kubectl"

// NOTE: versioned kubectl filenames must NOT start with "kubectl-", since
// that is reserved for plugins. Therefore, we prefix versioned kubectl
// filenames with "kubectl.". Example: "kubectl.1.12"
func createKubectlBinaryFilename(major string, minor string) (string, error) {
	return fmt.Sprintf("%s.%s.%s", kubectlBinaryName, major, minor), nil
}

// Example:
//   9+ -> 9
//   9.3 -> 9
//   9.1-gke -> 9
func normalizeVersionStr(majorMinor string) (string, error) {
	trimmed := strings.TrimSpace(majorMinor)
	if trimmed == "" {
		return "", fmt.Errorf("Empty server version major/minor string")
	}
	versionStr := ""
	for _, c := range trimmed {
		if unicode.IsDigit(c) {
			versionStr += string(c)
		} else {
			break
		}
	}
	if versionStr == "" {
		return "", fmt.Errorf("Bad server version major/minor string (%s)", trimmed)
	}
	return versionStr, nil
}

func isPositiveInteger(str string) bool {
	i, err := strconv.Atoi(str)
	if err != nil || i <= 0 { // NOTE: zero is also not allowed
		return false
	}
	return true
}
