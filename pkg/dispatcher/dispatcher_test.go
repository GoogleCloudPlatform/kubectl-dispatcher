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
	"testing"

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var clientVersion = &version.Info{
	Major:      "1",
	Minor:      "11",
	GitVersion: "v1.11.7",
}

func TestGetArgs(t *testing.T) {
	tests := []struct {
		args     []string
		remove   []string
		expected []string
	}{
		{
			args: []string{"foo", "bar"},
		},
		{
			args: []string{},
		},
	}
	for _, test := range tests {
		dispatcher := NewDispatcher(test.args, []string{}, clientVersion, nil)
		actual := dispatcher.GetArgs()
		if !isStringSliceEqual(test.args, actual) {
			t.Errorf("Filter args error: expected (%v), got (%v)", test.args, actual)
		}
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		env []string
	}{
		{
			env: []string{"ENV1", "ENV2", "ENV3"},
		},
		{
			env: []string{},
		},
	}
	for _, test := range tests {
		dispatcher := NewDispatcher([]string{}, test.env, clientVersion, nil)
		actual := dispatcher.GetEnv()
		if !isStringSliceEqual(test.env, actual) {
			t.Errorf("GetEnv() error: expected (%v), got (%v)", test.env, actual)
		}
	}
}

func isStringSliceEqual(a, b []string) bool {
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

func TestGetClientVersion(t *testing.T) {
	dispatcher := NewDispatcher([]string{}, []string{}, clientVersion, nil)
	actual := dispatcher.GetClientVersion()
	if clientVersion != actual {
		t.Errorf("GetClientVersion() error: expected (%v), got (%v)", clientVersion, actual)
	}
}

func TestInitKubeConfigFlags(t *testing.T) {
	tests := []struct {
		args  map[string]string
		flags *genericclioptions.ConfigFlags
	}{
		{
			args: map[string]string{
				"--cluster": "fake-cluster-name",
			},
		},
		{
			args: map[string]string{
				"--cluster":         "fake-cluster-name",
				"--server":          "https://127.0.0.1",
				"--context":         "fake-context",
				"--namespace":       "fake-namespace",
				"--request-timeout": "5",
			},
		},
		{
			args: map[string]string{
				"--cluster":         "fake-cluster-name",
				"--server":          "https://127.0.0.1",
				"--context":         "fake-context",
				"--extra-flag-1":    "not-used",
				"--extra-flag-2":    "not-used",
				"--namespace":       "fake-namespace",
				"--request-timeout": "5",
				"--extra-flag-3":    "not-used",
			},
		},
		{
			args: map[string]string{},
		},
	}
	for _, test := range tests {
		dispatcher := NewDispatcher(argsListFromMap(test.args), []string{}, nil, nil)
		expected := createConfigFlags(test.args)
		actual, err := dispatcher.InitKubeConfigFlags()
		if err != nil {
			t.Errorf("Unexpected error in InitKubeConfigFlags(): %v", err)
		}
		compareConfigFlags(t, expected, actual)
	}
}

// argsListFromMap creates a list of arguments from a map,
// prepending the "kubectl" command to the beginning of the list.
func argsListFromMap(args map[string]string) []string {
	flattened := []string{"kubectl"}
	for key, value := range args {
		flattened = append(flattened, key, value)
	}
	return flattened
}

// This must be kept in sync with "compareConfigFlags()"
func createConfigFlags(args map[string]string) *genericclioptions.ConfigFlags {
	flags := genericclioptions.NewConfigFlags(true)
	if value, ok := args["--cluster"]; ok {
		*flags.ClusterName = value
	}
	if value, ok := args["--context"]; ok {
		*flags.Context = value
	}
	if value, ok := args["--namespace"]; ok {
		*flags.Namespace = value
	}
	if value, ok := args["--server"]; ok {
		*flags.APIServer = value
	}
	if value, ok := args["--request-timeout"]; ok {
		*flags.Timeout = value
	}
	return flags
}

// Compares two ConfigFlag structs. Only compares a subset of fields. Must be kept
// in sync with "createConfigFlags()".
func compareConfigFlags(t *testing.T,
	expected *genericclioptions.ConfigFlags,
	actual *genericclioptions.ConfigFlags) bool {

	isEqual := true
	if *expected.ClusterName != *actual.ClusterName {
		t.Errorf("ConfigFlag Error: expected cluster name (%s), got (%s)", *expected.ClusterName, *actual.ClusterName)
		isEqual = false
	}
	if *expected.Context != *actual.Context {
		t.Errorf("ConfigFlag Error: expected context (%s), got (%s)", *expected.Context, *actual.Context)
		isEqual = false
	}
	if *expected.Namespace != *actual.Namespace {
		t.Errorf("ConfigFlag Error: expected namespace (%s), got (%s)", *expected.Namespace, *actual.Namespace)
		isEqual = false
	}
	if *expected.APIServer != *actual.APIServer {
		t.Errorf("ConfigFlag Error: expected server (%s), got (%s)", *expected.APIServer, *actual.APIServer)
		isEqual = false
	}
	if *expected.Timeout != *actual.Timeout {
		t.Errorf("ConfigFlag Error: expected timeout (%s), got (%s)", *expected.Timeout, *actual.Timeout)
		isEqual = false
	}

	return isEqual
}
