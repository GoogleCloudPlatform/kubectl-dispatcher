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

package util

import (
	"testing"

	"k8s.io/apimachinery/pkg/version"
)

func TestFilterList(t *testing.T) {
	tests := []struct {
		args     []string
		toRemove []string
		expected []string
	}{
		{
			args:     []string{},
			toRemove: []string{"-h"},
			expected: []string{},
		},
		{
			args:     []string{"foo"},
			toRemove: []string{"-h"},
			expected: []string{"foo"},
		},
		{
			args:     []string{"-h"},
			toRemove: []string{"-h"},
			expected: []string{},
		},
		{
			args:     []string{"-h", "-h"},
			toRemove: []string{"-h"},
			expected: []string{},
		},
		{
			args:     []string{"-h", "bar", "-h", "-h"},
			toRemove: []string{"-h"},
			expected: []string{"bar"},
		},
		{
			args:     []string{"foo", "-h", "bar"},
			toRemove: []string{"-h"},
			expected: []string{"foo", "bar"},
		},
		{
			args:     []string{"foo", "-h", "bar", "-h"},
			toRemove: []string{"-h"},
			expected: []string{"foo", "bar"},
		},
	}
	for _, test := range tests {
		actual := FilterList(test.args, test.toRemove)
		if !slicesEqual(actual, test.expected) {
			t.Errorf("Expected args (%v), got (%v)", test.expected, actual)
		}
	}
}

func slicesEqual(a, b []string) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestVersionMatch(t *testing.T) {
	tests := []struct {
		v1          version.Info
		v2          version.Info
		expectEqual bool
	}{
		{
			v1: version.Info{
				Major:      "1",
				Minor:      "11",
				GitVersion: "v1.11.7",
			},
			v2: version.Info{
				Major:      "1",
				Minor:      "11",
				GitVersion: "v1.11.7",
			},
			expectEqual: true,
		},
		{
			v1: version.Info{
				Major:      "1",
				Minor:      "10",
				GitVersion: "v1.10.7",
			},
			v2: version.Info{
				Major:      "1",
				Minor:      "11",
				GitVersion: "v1.11.7",
			},
			expectEqual: false,
		},
		{
			v1: version.Info{
				Major:      "1",
				Minor:      "11",
				GitVersion: "v1.11.7",
			},
			v2: version.Info{
				Major:      "1",
				Minor:      "11+",
				GitVersion: "v1.11.7",
			},
			expectEqual: true,
		},
		{
			v1: version.Info{
				Major:      "2",
				Minor:      "11",
				GitVersion: "v2.11.7",
			},
			v2: version.Info{
				Major:      "1",
				Minor:      "11",
				GitVersion: "v1.11.7",
			},
			expectEqual: false,
		},
		{
			v1: version.Info{
				Major:      "foo",
				Minor:      "bar",
				GitVersion: "v2.11.7",
			},
			v2: version.Info{
				Major:      "1",
				Minor:      "11",
				GitVersion: "v1.11.7",
			},
			expectEqual: false,
		},
		{
			v1: version.Info{
				Major:      "",
				Minor:      "",
				GitVersion: "",
			},
			v2: version.Info{
				Major:      "1",
				Minor:      "11",
				GitVersion: "v1.11.7",
			},
			expectEqual: false,
		},
	}
	for _, test := range tests {
		actual := VersionMatch(test.v1, test.v2)
		if test.expectEqual != actual {
			t.Errorf("VersionMatch error: expected (%t), got (%t) for (%+v)/(%+v)", test.expectEqual, actual, test.v1, test.v2)
		}
	}
}
