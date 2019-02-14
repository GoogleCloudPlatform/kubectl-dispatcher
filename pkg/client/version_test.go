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
	"testing"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestNewServerVersionClient(t *testing.T) {
	c := NewServerVersionClient(genericclioptions.NewConfigFlags(true))
	if c == nil {
		t.Errorf("New server version client returned nil")
	}
	if c.flags == nil {
		t.Errorf("New server version client did not populate flags field")
	}
	if c.delegate != nil {
		t.Errorf("New server version client improperly populated delegate field")
	}
	if c.requestTimeout != defaultRequestTimeout {
		t.Errorf("New server version client did not populate request timeout with default")
	}
	if c.cacheMaxAge != defaultCacheMaxAge {
		t.Errorf("New server version client did not populate cacheMaxAge with default")
	}
}

func TestRequestTimeout(t *testing.T) {
	svclient := NewServerVersionClient(nil)
	actual := svclient.GetRequestTimeout()
	if actual != defaultRequestTimeout {
		t.Errorf("Default request timeout: expected (%s), got (%s)", defaultRequestTimeout, actual)
	}
	const timeoutStr = "160ms"
	err := svclient.SetRequestTimeout(timeoutStr)
	if err != nil {
		t.Errorf("Unexpected error setting request timeout: (%v)", err)
	}
	expected, _ := time.ParseDuration(timeoutStr)
	actual = svclient.GetRequestTimeout()
	if expected != actual {
		t.Errorf("Request timeout error: expected (%s), got (%s)", expected, actual)
	}
	err = svclient.SetRequestTimeout("foobar")
	if err == nil {
		t.Errorf("Expected error for invalid duration in SetRequestTimeout did not occur")
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
