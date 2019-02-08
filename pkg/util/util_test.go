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
)

func TestRemoveArg(t *testing.T) {
	tests := []struct {
		args     []string
		toRemove string
		expected []string
	}{
		{
			args:     []string{},
			toRemove: "-h",
			expected: []string{},
		},
		{
			args:     []string{"foo"},
			toRemove: "-h",
			expected: []string{"foo"},
		},
		{
			args:     []string{"-h"},
			toRemove: "-h",
			expected: []string{},
		},
		{
			args:     []string{"-h", "-h"},
			toRemove: "-h",
			expected: []string{},
		},
		{
			args:     []string{"-h", "bar", "-h", "-h"},
			toRemove: "-h",
			expected: []string{"bar"},
		},
		{
			args:     []string{"foo", "-h", "bar"},
			toRemove: "-h",
			expected: []string{"foo", "bar"},
		},
		{
			args:     []string{"foo", "-h", "bar", "-h"},
			toRemove: "-h",
			expected: []string{"foo", "bar"},
		},
	}
	for _, test := range tests {
		actual := RemoveAllElements(test.args, test.toRemove)
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
