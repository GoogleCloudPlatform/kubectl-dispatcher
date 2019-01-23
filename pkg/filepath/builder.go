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
	version   *version.Info
	dirGetter DirectoryGetter
}

func NewFilepathBuilder(serverVersion *version.Info, dirGetter DirectoryGetter) (*FilepathBuilder, error) {
	if serverVersion == nil {
		return nil, fmt.Errorf("Missing server version")
	}
	if dirGetter == nil {
		return nil, fmt.Errorf("Missing directory getter")
	}
	return &FilepathBuilder{
		version:   serverVersion,
		dirGetter: dirGetter,
	}, nil
}

// VersionedFilePath returns the full absolute file path (or an error) to the
// versioned kubectl binary to dispatch to.
func (c *FilepathBuilder) VersionedFilePath() (string, error) {
	majorVersion, err := c.getMajorVersion()
	if err != nil {
		return "", err
	}
	minorVersion, err := c.getMinorVersion()
	if err != nil {
		return "", err
	}
	// Example: major: "1", minor: "12" -> "kubectl-1.12"
	kubectlFilename, err := createKubectlBinaryFilename(majorVersion, minorVersion)
	if err != nil {
		return "", err
	}
	currentDirectory, err := c.dirGetter.CurrentDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(currentDirectory, kubectlFilename), nil
}

func (c *FilepathBuilder) getMajorVersion() (string, error) {
	majorStr, err := normalizeVersionStr(c.version.Major)
	if err != nil {
		return "", err
	}
	if !isPositiveInteger(majorStr) {
		return "", fmt.Errorf("Bad major version string: %s", majorStr)
	}
	return majorStr, nil
}

func (c *FilepathBuilder) getMinorVersion() (string, error) {
	minorStr, err := normalizeVersionStr(c.version.Minor)
	if err != nil {
		return "", err
	}
	if !isPositiveInteger(minorStr) {
		return "", fmt.Errorf("Bad minor version string: %s", minorStr)
	}
	return minorStr, nil
}

const kubectlBinaryName = "kubectl"

func createKubectlBinaryFilename(major string, minor string) (string, error) {
	return fmt.Sprintf("%s-%s.%s", kubectlBinaryName, major, minor), nil
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
