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

// FilterArgs returns a copy of "l" with elements from "toRemove" filtered out.
func FilterList(l []string, rl []string) []string {
	c := CopyStrSlice(l)
	for _, r := range rl {
		c = RemoveAllElements(c, r)
	}
	return c
}

// RemoveAllElements removes all elements from "s" which match the string "r".
func RemoveAllElements(s []string, r string) []string {
	for i, rlen := 0, len(s); i < rlen; i++ {
		j := i - (rlen - len(s))
		if s[j] == r {
			s = append(s[:j], s[j+1:]...)
		}
	}
	return s
}

// CopyStrSlice returns a copy of the slice of strings.
func CopyStrSlice(s []string) []string {
	c := make([]string, len(s))
	copy(c, s)
	return c
}
