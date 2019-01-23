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

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
)

// Encapsulates the client which fetches the server version. Implements
// the discovery.ServerVersionInterface, allowing the creation of a
// mock or fake for testing.
type ServerVersionClient struct {
	delegate discovery.ServerVersionInterface
	version  *version.Info
}

func NewServerVersionClient(client discovery.ServerVersionInterface) (*ServerVersionClient, error) {
	if client == nil {
		return nil, fmt.Errorf("Missing discovery client")
	}
	return &ServerVersionClient{
		delegate: client,
		version:  nil,
	}, nil
}

// ServerVersion returns the (possibly cached) server version, using the
// stored discovery client.
func (c *ServerVersionClient) ServerVersion() (*version.Info, error) {
	if c.version == nil {
		version, err := c.delegate.ServerVersion()
		if err != nil {
			return nil, err
		}
		if version == nil {
			return nil, fmt.Errorf("Empty server version")
		}
		c.version = version
	}
	return c.version, nil
}
