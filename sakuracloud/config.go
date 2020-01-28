// Copyright 2016-2020 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sakuracloud

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/httpclient"
	"github.com/sacloud/libsacloud/api"
)

const uaEnvVar = "SAKURACLOUD_APPEND_USER_AGENT"

// Config type of SakuraCloud Config
type Config struct {
	AccessToken         string
	AccessTokenSecret   string
	Zone                string
	TimeoutMinute       int
	TraceMode           bool
	AcceptLanguage      string
	APIRootURL          string
	RetryMax            int
	RetryInterval       int
	APIRequestTimeout   int
	APIRequestRateLimit int
	terraformVersion    string
}

// APIClient for SakuraCloud API
type APIClient struct {
	*api.Client
}

// NewClient returns new API Client for SakuraCloud
func (c *Config) NewClient() *APIClient {
	client := api.NewClient(c.AccessToken, c.AccessTokenSecret, c.Zone)

	if c.AcceptLanguage != "" {
		client.AcceptLanguage = c.AcceptLanguage
	}
	if c.APIRootURL != "" {
		api.SakuraCloudAPIRoot = c.APIRootURL
	}
	if c.RetryMax > 0 {
		client.RetryMax = c.RetryMax
	}
	if c.RetryInterval > 0 {
		client.RetryInterval = time.Duration(c.RetryInterval) * time.Second
	}
	if c.TimeoutMinute > 0 {
		client.DefaultTimeoutDuration = time.Duration(c.TimeoutMinute) * time.Minute
	}

	httpClient := http.DefaultClient
	if c.APIRequestTimeout > 0 {
		httpClient.Timeout = time.Duration(c.APIRequestTimeout) * time.Second
	}
	if c.APIRequestRateLimit > 0 {
		httpClient.Transport = &api.RateLimitRoundTripper{RateLimitPerSec: c.APIRequestRateLimit}
	}
	client.HTTPClient = httpClient

	if c.TraceMode {
		client.TraceMode = true
		log.SetPrefix("[DEBUG] ")
	}
	tfUserAgent := httpclient.TerraformUserAgent(c.terraformVersion)
	providerUserAgent := fmt.Sprintf("%s/v%s", "terraform-provider-sakuracloud", Version)
	ua := fmt.Sprintf("%s %s", tfUserAgent, providerUserAgent)
	if add := os.Getenv(uaEnvVar); add != "" {
		ua += " " + add
		log.Printf("[DEBUG] Using modified User-Agent: %s", ua)
	}
	client.UserAgent = ua

	return &APIClient{
		Client: client,
	}
}
