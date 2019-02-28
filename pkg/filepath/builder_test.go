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
	"testing"

	"k8s.io/apimachinery/pkg/version"
)

type FakeDirGetter struct {
	dir string
	err error
}

func (f FakeDirGetter) CurrentDirectory() (string, error) {
	return f.dir, f.err
}

func createFakeDirGetter(dir string, err error) FakeDirGetter {
	return FakeDirGetter{
		dir: dir,
		err: err,
	}
}

func createServerVersion(major string, minor string) *version.Info {
	return &version.Info{
		Major:      major,
		Minor:      minor,
		GitVersion: "SHOULD BE UNUSED",
	}
}

func TestDefaultFilepath(t *testing.T) {
	tests := []struct {
		dirGetter DirectoryGetter
		filePath  string
	}{
		{
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar/kubectl.1.11",
		},
		{
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: fmt.Errorf("force dir error")},
			filePath:  "kubectl.1.11",
		},
	}

	defaultVersion := "1.11"
	for _, test := range tests {
		builder := NewFilepathBuilder(test.dirGetter, nil)
		filePath := builder.DefaultFilePath(defaultVersion)
		if filePath != test.filePath {
			t.Errorf("Expected default file path (%s), got (%s)", test.filePath, filePath)
		}
	}
}

func TestVersionedFilepath(t *testing.T) {
	tests := []struct {
		version   *version.Info
		dirGetter DirectoryGetter
		filePath  string
	}{
		{
			version:   createServerVersion("1", "9"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar/kubectl.1.9",
		},
		{
			version:   createServerVersion("\t1", " 9\n"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar/kubectl.1.9",
		},
		{
			version:   createServerVersion("1", "9+"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar/kubectl.1.9",
		},
		{
			version:   createServerVersion("1", "9.9-gke.1"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar/kubectl.1.9",
		},
		{
			version:   createServerVersion("1", "12"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar/kubectl.1.12",
		},
		{
			version:   createServerVersion("\t1", " 12\n"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar/kubectl.1.12",
		},
		{
			version:   createServerVersion("1", "12+"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar/kubectl.1.12",
		},
		{
			version:   createServerVersion("1", "12.3-gke.1"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar/kubectl.1.12",
		},
		// Nil server version maps to default kubectl
		{
			version:   nil,
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar",
		},
		// Non-digit major version not allowed
		{
			version:   createServerVersion("k", "9"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar",
		},
		// Non-digit minor version not allowed
		{
			version:   createServerVersion("1", "s"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar",
		},
		// Empty/space major version not allowed
		{
			version:   createServerVersion(" \t", "9"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar",
		},
		// Empty/space minor version not allowed
		{
			version:   createServerVersion("1", "\n"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar",
		},
		// Zero as major version not allowed
		{
			version:   createServerVersion("0", "9"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar",
		},
		// Zero as minor version not allowed
		{
			version:   createServerVersion("1", "0"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:  "/foo/bar",
		},
		// Nil directory getter defaults to no directory
		{
			version:   createServerVersion("1", "9"),
			dirGetter: nil,
			filePath:  "kubectl.1.9",
		},
		// // Error in retrieving current directory defaults to no directory
		{
			version:   createServerVersion("1", "9"),
			dirGetter: FakeDirGetter{dir: "/foo/bar", err: fmt.Errorf("Forced error")},
			filePath:  "kubectl.1.9",
		},
	}
	for _, test := range tests {
		builder := NewFilepathBuilder(test.dirGetter, nil)
		filePath := builder.VersionedFilePath(test.version)
		if filePath != test.filePath {
			t.Errorf("Expected versioned file path (%s), got (%s)", test.filePath, filePath)
		}
	}
}

func FilepathIsValid(string) (os.FileInfo, error) {
	return nil, nil
}

func FilepathIsNotValid(string) (os.FileInfo, error) {
	return nil, fmt.Errorf("Forced filepath validate error")
}

func TestValidateFilepath(t *testing.T) {
	tests := []struct {
		dirGetter    DirectoryGetter
		validateFunc func(string) (os.FileInfo, error)
		expectError  bool
	}{
		{
			dirGetter:    FakeDirGetter{dir: "/foo/bar", err: nil},
			validateFunc: FilepathIsValid,
			expectError:  false,
		},
		{
			dirGetter:    FakeDirGetter{dir: "/foo/bar", err: nil},
			validateFunc: FilepathIsNotValid,
			expectError:  true,
		},
	}

	for _, test := range tests {
		builder := NewFilepathBuilder(test.dirGetter, test.validateFunc)
		if !test.expectError {
			if err := builder.ValidateFilepath("doesn't matter"); err != nil {
				t.Errorf("Unexpected error in ValidateFilepath")
			}
		} else {
			if err := builder.ValidateFilepath("doesn't matter"); err == nil {
				t.Errorf("Expected error in ValidateFilepath not received")
			}
		}
	}
}
