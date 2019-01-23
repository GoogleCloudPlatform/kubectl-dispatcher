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

func TestNewServerVersionClient(t *testing.T) {
	fake := FakeDiscoveryClient{}
	client, err := NewServerVersionClient(fake)
	if client == nil {
		t.Errorf("Error: missing server version client")
	}
	if err != nil {
		t.Errorf("Unexpected error creating server version client: (%v)", err)
	}
	if client.delegate != fake {
		t.Errorf("Server version client: delegate not properly initialized")
	}
	if client.version != nil {
		t.Errorf("Server version client: version field is unexpectedly initialized (%v)", client.version)
	}
	// Validate that a nil parameter causes an error
	client, err = NewServerVersionClient(nil)
	if client != nil || err == nil {
		t.Errorf("Nil discovery client parameter should return error")
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
			version:      &version.Info{Major: "1", Minor: "10"},
			discoveryErr: fmt.Errorf("Force discovery client error"),
			expectedErr:  true,
		},
		// Empty returned server version should cause error
		{
			version:      nil,
			discoveryErr: nil,
			expectedErr:  true,
		},
	}
	for _, test := range tests {
		client, _ := NewServerVersionClient(createFakeDiscoveryClient(test.version, test.discoveryErr))
		version, err := client.ServerVersion()
		if err != nil && !test.expectedErr {
			t.Errorf("Unexpected error in versioned file path: (%v)", err)
		}
		if test.expectedErr {
			if err == nil {
				t.Errorf("Expected error did not occur fetching ServerVersion")
			}
		} else {
			if version != test.version {
				t.Errorf("Expected server version (%s), got (%s)", test.version, version)
			}
			if client.version == nil {
				t.Errorf("Server version client should have cached the server version")
			}
		}
	}
}
