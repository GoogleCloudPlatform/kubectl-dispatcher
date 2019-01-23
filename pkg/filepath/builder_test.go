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

func TestCreateFilepathBuilder(t *testing.T) {
	tests := []struct {
		version     *version.Info
		dirGetter   DirectoryGetter
		expectedErr bool
	}{
		{
			version:     createServerVersion("1", "9"),
			dirGetter:   FakeDirGetter{},
			expectedErr: false,
		},
		{
			version:     nil,
			dirGetter:   FakeDirGetter{},
			expectedErr: true,
		},
		{
			version:     createServerVersion("1", "10"),
			dirGetter:   nil,
			expectedErr: true,
		},
	}
	for _, test := range tests {
		_, err := NewFilepathBuilder(test.version, test.dirGetter)
		if err != nil && !test.expectedErr {
			t.Errorf("Unexpected error creating filepath builder: (%v)", err)
		}
		if err == nil && test.expectedErr {
			t.Errorf("Expected error did not occur creating filepath builder")
		}
	}
}

func TestVersionedFilepath(t *testing.T) {
	tests := []struct {
		version     *version.Info
		dirGetter   DirectoryGetter
		filePath    string
		expectedErr bool
	}{
		{
			version:     createServerVersion("1", "9"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl-1.9",
			expectedErr: false,
		},
		{
			version:     createServerVersion("\t1", " 9\n"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl-1.9",
			expectedErr: false,
		},
		{
			version:     createServerVersion("1", "9+"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl-1.9",
			expectedErr: false,
		},
		{
			version:     createServerVersion("1", "9.9-gke.1"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl-1.9",
			expectedErr: false,
		},
		{
			version:     createServerVersion("1", "12"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl-1.12",
			expectedErr: false,
		},
		{
			version:     createServerVersion("\t1", " 12\n"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl-1.12",
			expectedErr: false,
		},
		{
			version:     createServerVersion("1", "12+"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl-1.12",
			expectedErr: false,
		},
		{
			version:     createServerVersion("1", "12.3-gke.1"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "/foo/bar/kubectl-1.12",
			expectedErr: false,
		},
		// Error: Non-digit major version not allowed
		{
			version:     createServerVersion("k", "9"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "NOT USED",
			expectedErr: true,
		},
		// Error: Non-digit minor version not allowed
		{
			version:     createServerVersion("1", "s"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "NOT USED",
			expectedErr: true,
		},
		// Error: Empty/space major version not allowed
		{
			version:     createServerVersion(" \t", "9"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "NOT USED",
			expectedErr: true,
		},
		// Error: Empty/space minor version not allowed
		{
			version:     createServerVersion("1", "\n"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "NOT USED",
			expectedErr: true,
		},
		// Error: Zero as major version not allowed
		{
			version:     createServerVersion("0", "9"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "NOT USED",
			expectedErr: true,
		},
		// Error: Zero as minor version not allowed
		{
			version:     createServerVersion("1", "0"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: nil},
			filePath:    "NOT USED",
			expectedErr: true,
		},
		// Force directory getter error
		{
			version:     createServerVersion("1", "9"),
			dirGetter:   FakeDirGetter{dir: "/foo/bar", err: fmt.Errorf("Forced error")},
			filePath:    "NOT USED",
			expectedErr: true,
		},
	}
	for _, test := range tests {
		builder, _ := NewFilepathBuilder(test.version, test.dirGetter)
		filePath, err := builder.VersionedFilePath()
		if err != nil && !test.expectedErr {
			t.Errorf("Unexpected error in versioned file path: (%v)", err)
		}
		if test.expectedErr {
			if err == nil {
				t.Errorf("Expected error did not occur creating versioned file path")
			}
		} else if filePath != test.filePath {
			t.Errorf("Expected versioned file path (%s), got (%s)", test.filePath, filePath)
		}
	}
}
