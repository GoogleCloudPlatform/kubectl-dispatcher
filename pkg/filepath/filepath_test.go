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
	os  string
	dir string
	err error
}

func (f FakeDirGetter) CurrentDirectory() (string, error) {
	return f.dir, f.err
}

func (f FakeDirGetter) GetOS() string {
	return f.os
}

func createServerVersion(major string, minor string) version.Info {
	return version.Info{
		Major:      major,
		Minor:      minor,
		GitVersion: "SHOULD BE UNUSED",
	}
}

func TestVersionedFilepath(t *testing.T) {
	tests := []struct {
		version     version.Info
		dirGetter   DirectoryGetter
		filePath    string
		expectError bool
	}{
		{
			version:     createServerVersion("1", "9"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl.1.9",
			expectError: false,
		},
		{
			version:     createServerVersion("1", "9"),
			dirGetter:   FakeDirGetter{os: "darwin", dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl.1.9",
			expectError: false,
		},
		{
			version:     createServerVersion("1", "9"),
			dirGetter:   FakeDirGetter{os: "windows", dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl.1.9.exe",
			expectError: false,
		},
		{
			version:     createServerVersion("\t1", " 9\n"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl.1.9",
			expectError: false,
		},
		{
			version:     createServerVersion("1", "9+"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl.1.9",
			expectError: false,
		},
		{
			version:     createServerVersion("1", "9.9-gke.1"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl.1.9",
			expectError: false,
		},
		{
			version:     createServerVersion("1", "12"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl.1.12",
			expectError: false,
		},
		{
			version:     createServerVersion("\t1", " 12\n"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl.1.12",
			expectError: false,
		},
		{
			version:     createServerVersion("1", "12+"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl.1.12",
			expectError: false,
		},
		{
			version:     createServerVersion("1", "12.3-gke.1"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl.1.12",
			expectError: false,
		},
		// Non-digit major version not allowed
		{
			version:     createServerVersion("k", "9"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "",
			expectError: true,
		},
		// Non-digit minor version not allowed
		{
			version:     createServerVersion("1", "s"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "",
			expectError: true,
		},
		// Empty/space major version not allowed
		{
			version:     createServerVersion(" \t", "9"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "",
			expectError: true,
		},
		// Empty/space minor version not allowed
		{
			version:     createServerVersion("1", "\n"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "",
			expectError: true,
		},
		// Zero as major version not allowed
		{
			version:     createServerVersion("0", "9"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "",
			expectError: true,
		},
		// Zero as minor version not allowed
		{
			version:     createServerVersion("1", "0"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			filePath:    "",
			expectError: true,
		},
		// Nil directory getter returns error.
		{
			version:     createServerVersion("1", "9"),
			dirGetter:   nil,
			filePath:    "",
			expectError: true,
		},
		// Error forced in retrieving current directory.
		{
			version:     createServerVersion("1", "9"),
			dirGetter:   FakeDirGetter{os: "linux", dir: "", err: fmt.Errorf("Forced error")},
			filePath:    "",
			expectError: true,
		},
	}
	for _, test := range tests {
		builder := NewFilepathBuilder(test.dirGetter, nil)
		filePath, err := builder.VersionedFilePath(test.version)
		if !test.expectError {
			if err != nil {
				t.Errorf("Unexpected error: (%v)", err)
			} else {
				if filePath != test.filePath {
					t.Errorf("Expected versioned file path (%s), got (%s)", test.filePath, filePath)
				}
			}
		}
		if test.expectError && (err == nil) {
			t.Errorf("Expected error; received none")
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
			dirGetter:    FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
			validateFunc: FilepathIsValid,
			expectError:  false,
		},
		{
			dirGetter:    FakeDirGetter{os: "linux", dir: "/foo/bar", err: nil},
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
