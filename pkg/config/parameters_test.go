// Copyright Project Contour Authors
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

package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestGetenvOr(t *testing.T) {
	assert.Equal(t, t.Name(), GetenvOr("B5E09AAD-DEFC-4650-9DE6-0F2E3AF7FCF2", t.Name()))

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		assert.NotEqual(t, t.Name(), GetenvOr(parts[0], t.Name()))
	}
}

func TestParseDefaults(t *testing.T) {
	savedHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", savedHome)
	}()

	require.NoError(t, os.Setenv("HOME", t.Name()))

	data, err := yaml.Marshal(Defaults())
	require.NoError(t, err)

	expected := `
debug: false
kubeconfig: TestParseDefaults/.kube/config
server:
  xds-server-type: contour
accesslog-format: envoy
json-fields:
- '@timestamp'
- authority
- bytes_received
- bytes_sent
- downstream_local_address
- downstream_remote_address
- duration
- method
- path
- protocol
- request_id
- requested_server_name
- response_code
- response_flags
- uber_trace_id
- upstream_cluster
- upstream_host
- upstream_local_address
- upstream_service_time
- user_agent
- x_forwarded_for
leaderelection:
  lease-duration: 15s
  renew-deadline: 10s
  retry-period: 2s
  configmap-namespace: projectcontour
  configmap-name: leader-elect
timeouts:
  connection-idle-timeout: 60s
envoy-service-namespace: projectcontour
envoy-service-name: envoy
default-http-versions: []
cluster:
  dns-lookup-family: auto
`
	assert.Equal(t, strings.TrimSpace(string(data)), strings.TrimSpace(expected))

	conf, err := Parse(strings.NewReader(expected))
	require.NoError(t, err)
	require.NoError(t, conf.Validate())

	wanted := Defaults()
	assert.Equal(t, &wanted, conf)
}

func TestParseFailure(t *testing.T) {
	badYAML := `
foo: bad

`
	_, err := Parse(strings.NewReader(badYAML))
	require.Error(t, err)
}

func TestValidateClusterDNSFamilyType(t *testing.T) {
	assert.Error(t, ClusterDNSFamilyType("").Validate())
	assert.Error(t, ClusterDNSFamilyType("foo").Validate())

	assert.NoError(t, AutoClusterDNSFamily.Validate())
	assert.NoError(t, IPv4ClusterDNSFamily.Validate())
	assert.NoError(t, IPv6ClusterDNSFamily.Validate())
}

func TestValidateNamespacedName(t *testing.T) {
	assert.NoErrorf(t, NamespacedName{}.Validate(), "empty name should be OK")
	assert.NoError(t, NamespacedName{Name: "name", Namespace: "ns"}.Validate())

	assert.Error(t, NamespacedName{Name: "name"}.Validate())
	assert.Error(t, NamespacedName{Namespace: "ns"}.Validate())
}

func TestValidateServerType(t *testing.T) {
	assert.Error(t, ServerType("").Validate())
	assert.Error(t, ServerType("foo").Validate())

	assert.NoError(t, EnvoyServerType.Validate())
	assert.NoError(t, ContourServerType.Validate())
}

func TestValidateAccessLogType(t *testing.T) {
	assert.Error(t, AccessLogType("").Validate())
	assert.Error(t, AccessLogType("foo").Validate())

	assert.NoError(t, EnvoyAccessLog.Validate())
	assert.NoError(t, JSONAccessLog.Validate())
}

func TestValidateAccessLogFields(t *testing.T) {
	errorCases := [][]string{
		{"dog", "cat"},
		{"req"},
		{"resp"},
		{"trailer"},
		{"@timestamp", "dog"},
		{"@timestamp", "content-id=%REQ=dog%"},
		{"@timestamp", "content-id=%dog(%"},
		{"@timestamp", "content-id=%REQ()%"},
		{"@timestamp", "content-id=%DOG%"},
		{"@timestamp", "duration=my durations % are %DURATION%.0 and %REQ(:METHOD)%"},
		{"invalid=%REQ%"},
		{"invalid=%TRAILER%"},
		{"invalid=%RESP%"},
		{"@timestamp", "invalid=%START_TIME(%s.%6f):10%"},
	}

	for _, c := range errorCases {
		assert.Error(t, AccessLogFields(c).Validate(), c)
	}

	successCases := [][]string{
		{"@timestamp", "method"},
		{"start_time"},
		{"@timestamp", "response_duration"},
		{"@timestamp", "duration=%DURATION%.0"},
		{"@timestamp", "duration=My duration=%DURATION%.0"},
		{"@timestamp", "duratin=%START_TIME(%s.%6f)%"},
		{"@timestamp", "content-id=%REQ(X-CONTENT-ID)%"},
		{"@timestamp", "content-id=%REQ(X-CONTENT-ID):10%"},
		{"@timestamp", "length=%RESP(CONTENT-LENGTH):10%"},
		{"@timestamp", "trailer=%TRAILER(CONTENT-LENGTH):10%"},
		{"@timestamp", "duration=my durations are %DURATION%.0 and method is %REQ(:METHOD)%"},
		{"dog=pug", "cat=black"},
	}

	for _, c := range successCases {
		assert.NoError(t, AccessLogFields(c).Validate(), c)
	}
}

func TestValidateHTTPVersionType(t *testing.T) {
	assert.Error(t, HTTPVersionType("").Validate())
	assert.Error(t, HTTPVersionType("foo").Validate())
	assert.Error(t, HTTPVersionType("HTTP/1.1").Validate())
	assert.Error(t, HTTPVersionType("HTTP/2").Validate())

	assert.NoError(t, HTTPVersion1.Validate())
	assert.NoError(t, HTTPVersion2.Validate())
}

func TestValidateTimeoutParams(t *testing.T) {
	assert.NoError(t, TimeoutParameters{}.Validate())
	assert.NoError(t, TimeoutParameters{
		RequestTimeout:                "infinite",
		ConnectionIdleTimeout:         "infinite",
		StreamIdleTimeout:             "infinite",
		MaxConnectionDuration:         "infinite",
		ConnectionShutdownGracePeriod: "infinite",
	}.Validate())
	assert.NoError(t, TimeoutParameters{
		RequestTimeout:                "infinity",
		ConnectionIdleTimeout:         "infinity",
		StreamIdleTimeout:             "infinity",
		MaxConnectionDuration:         "infinity",
		ConnectionShutdownGracePeriod: "infinity",
	}.Validate())

	assert.Error(t, TimeoutParameters{RequestTimeout: "foo"}.Validate())
	assert.Error(t, TimeoutParameters{ConnectionIdleTimeout: "bar"}.Validate())
	assert.Error(t, TimeoutParameters{StreamIdleTimeout: "baz"}.Validate())
	assert.Error(t, TimeoutParameters{MaxConnectionDuration: "boop"}.Validate())
	assert.Error(t, TimeoutParameters{ConnectionShutdownGracePeriod: "bong"}.Validate())

}

func TestConfigFileValidation(t *testing.T) {
	check := func(yamlIn string) {
		t.Helper()

		conf, err := Parse(strings.NewReader(yamlIn))
		require.NoError(t, err)
		require.Error(t, conf.Validate())
	}

	check(`
cluster:
  dns-lookup-family: stone
`)

	check(`
server:
  xds-server-type: magic
`)

	check(`
accesslog-format: /dev/null
`)

	check(`
json-fields:
- one
`)

	check(`
tls:
  fallback-certificate:
    name: foo
`)

	check(`
tls:
  envoy-client-certificate:
    name: foo
`)

	check(`
timeouts:
  request-timeout: none
`)

	check(`
default-http-versions:
- http/0.9
`)

}

func TestConfigFileDefaultOverrideImport(t *testing.T) {
	check := func(verifier func(*testing.T, *Parameters), yamlIn string) {
		t.Helper()

		conf, err := Parse(strings.NewReader(yamlIn))

		require.NoError(t, err)
		verifier(t, conf)
	}

	check(func(t *testing.T, conf *Parameters) {
		wanted := Defaults()
		assert.Equal(t, &wanted, conf)
	}, "")

	check(func(t *testing.T, conf *Parameters) {
		wanted := Defaults()
		assert.Equal(t, &wanted, conf)
	}, `
incluster: false
disablePermitInsecure: false
disableAllowChunkedLength: false
leaderelection:
  configmap-name: leader-elect
  configmap-namespace: projectcontour
  lease-duration: 15s
  renew-deadline: 10s
  retry-period: 2s
`,
	)

	check(func(t *testing.T, conf *Parameters) {
		wanted := Defaults()
		assert.Equal(t, &wanted, conf)
	}, `
tls:
`)

	check(func(t *testing.T, conf *Parameters) {
		assert.Equal(t, "1.2", conf.TLS.MinimumProtocolVersion)
	}, `
tls:
  minimum-protocol-version: 1.2
`)

	check(func(t *testing.T, conf *Parameters) {
		assert.Equal(t, "foo", conf.LeaderElection.Name)
		assert.Equal(t, "bar", conf.LeaderElection.Namespace)
	}, `
leaderelection:
  configmap-name: foo
  configmap-namespace: bar
`)

	check(func(t *testing.T, conf *Parameters) {
		assert.Equal(t, conf.LeaderElection,
			LeaderElectionParameters{
				Name:          "foo",
				Namespace:     "bar",
				LeaseDuration: 600 * time.Second,
				RenewDeadline: 500 * time.Second,
				RetryPeriod:   60 * time.Second,
			})
	}, `
leaderelection:
  configmap-name: foo
  configmap-namespace: bar
  lease-duration: 600s
  renew-deadline: 500s
  retry-period: 60s
`)
	check(func(t *testing.T, conf *Parameters) {
		assert.ElementsMatch(t,
			[]HTTPVersionType{HTTPVersion1, HTTPVersion2, HTTPVersion2, HTTPVersion1},
			conf.DefaultHTTPVersions,
		)
	}, `
default-http-versions:
- http/1.1
- http/2
- HTTP/2
- HTTP/1.1
`)
}
