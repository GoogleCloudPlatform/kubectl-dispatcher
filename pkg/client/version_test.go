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
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	"k8s.io/client-go/rest/fake"
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

func TestServerVersionCorrectlyReturnsVersion(t *testing.T) {
	expected := createServerVersion(1, 10)
	serverVersionBytes, err := json.Marshal(*expected)
	if err != nil {
		t.Fatalf("Unexpected JSON marshal error for server version: (%v)", err)
	}
	var body io.ReadCloser = ioutil.NopCloser(bytes.NewReader(serverVersionBytes))
	fakeRestClient := &fake.RESTClient{
		NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch p, m := req.URL.Path, req.Method; {
			case p == "/version" && m == "GET":
				// Validate the Cache-Control header.
				if value, ok := req.Header["Cache-Control"]; !ok {
					t.Errorf("Missing Cache-Control HTTP header")
					if !strings.HasPrefix(value[0], "max-age=") {
						t.Errorf("Erroneous Cache-Control HTTP header: (%s)", value)
					}
				}
				// Validate the timeout query parameter was set.
				requestTimeout := req.URL.Query().Get("timeout")
				if requestTimeout != defaultRequestTimeout.String() {
					t.Errorf("Expected request timeout (%s), got (%s)", defaultRequestTimeout.String(), requestTimeout)
				}
				return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: body}, nil
			default:
				t.Fatalf("Unexpected request: %#v\n%#v", req.URL, req)
				return nil, nil
			}
		}),
	}
	svclient := NewServerVersionClient(nil) // kubeConfigFlags are unused
	svclient.delegate = fakeRestClient      // Force ServerVersionClient to use fake RESTClient
	actual, err := svclient.ServerVersion()
	if err != nil {
		t.Errorf("Unexpected error retrieving ServerVersion")
	}
	if expected.Major != actual.Major {
		t.Errorf("Expected server major version (%s), got (%s)", expected.Major, actual.Major)
	}
	if expected.Minor != actual.Minor {
		t.Errorf("Expected server minor version (%s), got (%s)", expected.Minor, actual.Minor)
	}
}

func TestServerVersionErrorOnBadVersion(t *testing.T) {
	// Set up an empty server version for the body of the HTTP response.
	var body io.ReadCloser = ioutil.NopCloser(bytes.NewReader([]byte{}))
	fakeRestClient := &fake.RESTClient{
		NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch p, m := req.URL.Path, req.Method; {
			case p == "/version" && m == "GET":
				// Validate the Cache-Control header.
				if value, ok := req.Header["Cache-Control"]; !ok {
					t.Errorf("Missing Cache-Control HTTP header")
					if !strings.HasPrefix(value[0], "max-age=") {
						t.Errorf("Erroneous Cache-Control HTTP header: (%s)", value)
					}
				}
				// Validate the timeout query parameter was set.
				requestTimeout := req.URL.Query().Get("timeout")
				if requestTimeout != defaultRequestTimeout.String() {
					t.Errorf("Expected request timeout (%s), got (%s)", defaultRequestTimeout.String(), requestTimeout)
				}
				return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: body}, nil
			default:
				t.Fatalf("Unexpected request: %#v\n%#v", req.URL, req)
				return nil, nil
			}
		}),
	}
	svclient := NewServerVersionClient(nil) // kubeConfigFlags are unused
	svclient.delegate = fakeRestClient      // Force ServerVersionClient to use fake RESTClient
	_, err := svclient.ServerVersion()
	if err == nil {
		t.Errorf("Expected error retrieving empty server version did not occur")
	}
}

func createServerVersion(major int, minor int) *version.Info {
	return &version.Info{
		Major: strconv.Itoa(major),
		Minor: strconv.Itoa(minor),
	}
}

func defaultHeader() http.Header {
	header := http.Header{}
	header.Set("Content-Type", runtime.ContentTypeJSON)
	return header
}
