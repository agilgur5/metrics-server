// Copyright 2020 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package options

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/metrics-server/pkg/scraper"
)

func TestKubeletConfig(t *testing.T) {
	kubeconfig := &rest.Config{
		Host:            "https://10.96.0.1:443",
		APIPath:         "",
		Username:        "Username",
		Password:        "Password",
		BearerToken:     "ApiserverBearerToken",
		BearerTokenFile: "ApiserverBearerTokenFile",
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false,
			CertFile: "CertFile",
			KeyFile:  "KeyFile",
			CAFile:   "CAFile",
			CertData: []byte("CertData"),
			KeyData:  []byte("KeyData"),
			CAData:   []byte("CAData"),
		},
		UserAgent: "UserAgent",
	}

	expected := scraper.KubeletClientConfig{
		AddressTypePriority: []v1.NodeAddressType{"Hostname", "InternalDNS", "InternalIP", "ExternalDNS", "ExternalIP"},
		Scheme:              "https",
		DefaultPort:         10250,
		Client:              *kubeconfig,
	}

	for _, tc := range []struct {
		name        string
		optionsFunc func() *Options
		expectFunc  func() scraper.KubeletClientConfig
		kubeconfig  *rest.Config
	}{
		{
			name: "Default configuration should use config from kubeconfig",
			optionsFunc: func() *Options {
				return NewOptions()
			},
			expectFunc: func() scraper.KubeletClientConfig {
				return expected
			},
		},
		{
			name: "InsecureKubeletTLS removes CA config and sets insecure",
			optionsFunc: func() *Options {
				o := NewOptions()
				o.InsecureKubeletTLS = true
				return o
			},
			expectFunc: func() scraper.KubeletClientConfig {
				e := expected
				e.Client.Insecure = true
				e.Client.CAFile = ""
				e.Client.CAData = nil
				return e
			},
		},
		{
			name: "KubeletCAFile overrides CA file and data",
			optionsFunc: func() *Options {
				o := NewOptions()
				o.KubeletCAFile = "Override"
				return o
			},
			expectFunc: func() scraper.KubeletClientConfig {
				e := expected
				e.Client.CAFile = "Override"
				e.Client.CAData = nil
				return e
			},
		},
		{
			name: "DeprecatedCompletelyInsecureKubelet resets TLSConfig and sets https scheme",
			optionsFunc: func() *Options {
				o := NewOptions()
				o.DeprecatedCompletelyInsecureKubelet = true
				return o
			},
			expectFunc: func() scraper.KubeletClientConfig {
				e := expected
				e.Client.TLSClientConfig = rest.TLSClientConfig{}
				e.Client.Username = ""
				e.Client.Password = ""
				e.Client.BearerToken = ""
				e.Client.BearerTokenFile = ""
				e.Scheme = "http"
				return e
			},
		},
		{
			name: "KubeletClientCertFile overrides TLS client cert file",
			optionsFunc: func() *Options {
				o := NewOptions()
				o.KubeletClientCertFile = "Override"
				return o
			},
			expectFunc: func() scraper.KubeletClientConfig {
				e := expected
				e.Client.TLSClientConfig.CertFile = "Override"
				e.Client.TLSClientConfig.CertData = nil
				return e
			},
		},
		{
			name: "KubeletClientKeyFile overrides TLS client key file",
			optionsFunc: func() *Options {
				o := NewOptions()
				o.KubeletClientKeyFile = "Override"
				return o
			},
			expectFunc: func() scraper.KubeletClientConfig {
				e := expected
				e.Client.TLSClientConfig.KeyFile = "Override"
				e.Client.TLSClientConfig.KeyData = nil
				return e
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.optionsFunc().kubeletConfig(kubeconfig)
			if diff := cmp.Diff(*config, tc.expectFunc()); diff != "" {
				t.Errorf("Unexpected options.KubeletConfig(), diff:\n%s", diff)
			}
		})
	}
}
