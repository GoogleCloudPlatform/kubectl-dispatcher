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

package client

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/version"
)

// FakeDiscoveryClient implements discovery.ServerVersionInterface.
type FakeDiscoveryClient struct {
	version *version.Info
	err     error
}

func (f FakeDiscoveryClient) ServerVersion() (*version.Info, error) {
	return f.version, f.err
}

func createFakeDiscoveryClient(version *version.Info, err error) FakeDiscoveryClient {
	return FakeDiscoveryClient{
		version: version,
		err:     err,
	}
}

func TestServerVersion(t *testing.T) {
	tests := []struct {
		version      *version.Info
		discoveryErr error
		expectedErr  bool
	}{
		{
			version:      &version.Info{Major: "1", Minor: "10"},
			discoveryErr: nil,
			expectedErr:  false,
		},
		// Force a discovery client error, which should propogate
		{
			version:      nil,
			discoveryErr: fmt.Errorf("Force discovery client error"),
			expectedErr:  true,
		},
	}
	for _, test := range tests {
		client := NewServerVersionClient(nil)
		client.Delegate = createFakeDiscoveryClient(test.version, test.discoveryErr)
		version, err := client.ServerVersion()
		if test.expectedErr {
			if err == nil {
				t.Errorf("Expected error did not occur fetching ServerVersion")
			}
		} else {
			if version != test.version {
				t.Errorf("Expected server version (%s), got (%s)", test.version, version)
			}
		}
	}
}

func TestRequestTimeout(t *testing.T) {
	expected := "160ms"
	svclient := NewServerVersionClient(nil)
	svclient.SetRequestTimeout(expected)
	actual := svclient.GetRequestTimeout()
	if expected != actual {
		t.Errorf("Request timeout error: expected (%s), got (%s)", expected, actual)
	}
}

func TestCacheMaxAge(t *testing.T) {
	expected := uint64(44000)
	svclient := NewServerVersionClient(nil)
	svclient.SetCacheMaxAge(expected)
	actual := svclient.GetCacheMaxAge()
	if expected != actual {
		t.Errorf("Request timeout error: expected (%d), got (%d)", expected, actual)
	}
}
