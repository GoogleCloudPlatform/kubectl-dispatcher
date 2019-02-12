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
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
)

const defaultRequestTimeout = "2s"
const defaultCacheMaxAge = 60 * 60 * 24 // Seconds in one day

// Encapsulates the client which fetches the server version. Implements
// the discovery.ServerVersionInterface, allowing the creation of a
// mock or fake for testing.
type ServerVersionClient struct {
	Flags          *genericclioptions.ConfigFlags
	Delegate       discovery.ServerVersionInterface
	requestTimeout string // Examples: "650ms", "2s"
	cacheMaxAge    uint64 // Maximum cache age allowed in seconds
}

func NewServerVersionClient(kubeConfigFlags *genericclioptions.ConfigFlags) *ServerVersionClient {
	return &ServerVersionClient{
		Flags:          kubeConfigFlags,
		Delegate:       nil,
		requestTimeout: defaultRequestTimeout,
		cacheMaxAge:    defaultCacheMaxAge,
	}
}

func (c *ServerVersionClient) GetRequestTimeout() string {
	return c.requestTimeout
}

func (c *ServerVersionClient) SetRequestTimeout(requestTimeout string) {
	c.requestTimeout = requestTimeout
}

func (c *ServerVersionClient) GetCacheMaxAge() uint64 {
	return c.cacheMaxAge
}

func (c *ServerVersionClient) SetCacheMaxAge(cacheMaxAge uint64) {
	c.cacheMaxAge = cacheMaxAge
}

func (c *ServerVersionClient) ServerVersion() (*version.Info, error) {
	// TODO: Implement caching here.
	// Cache the discovery client if we haven't already.
	if c.Delegate == nil {
		// TODO: Update timeout into the kube config flags, then create the discovery client.
		discoveryClient, err := c.Flags.ToDiscoveryClient()
		if err != nil {
			return nil, err
		}
		c.Delegate = discoveryClient
	}
	return c.Delegate.ServerVersion()
}
