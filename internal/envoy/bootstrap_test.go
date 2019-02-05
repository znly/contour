// Copyright © 2017 Heptio
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

package envoy

import (
	"testing"

	bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/google/go-cmp/cmp"
)

func TestBootstrap(t *testing.T) {
	tests := map[string]struct {
		config BootstrapConfig
		want   string
	}{
		"default configuration": {
			config: BootstrapConfig{},
			want: `{
  "static_resources": {
    "listeners": [
      {
        "address": {
          "socket_address": {
            "address": "0.0.0.0",
            "port_value": 8002
          }
        },
        "filter_chains": [
          {
            "filters": [
              {
                "name": "envoy.http_connection_manager",
                "config": {
                  "codec_type": "AUTO",
                  "http_filters": [
                    {
                      "config": {
                        "headers": [
                          {
                            "exact_match": "/healthz",
                            "name": ":path"
                          }
                        ],
                        "pass_through_mode": "false"
                      },
                      "name": "envoy.health_check"
                    },
                    {
                      "name": "envoy.router"
                    }
                  ],
                  "route_config": {
                    "virtual_hosts": {
                      "domains": [
                        "*"
                      ],
                      "name": "backend",
                      "routes": [
                        {
                          "match": {
                            "prefix": "/stats"
                          },
                          "route": {
                            "cluster": "service_stats"
                          }
                        }
                      ]
                    }
                  },
                  "stat_prefix": "stats"
                }
              }
            ]
          }
        ]
      }
    ],
    "clusters": [
      {
        "name": "contour",
        "type": "STRICT_DNS",
        "connect_timeout": "5s",
        "hosts": [
          {
            "socket_address": {
              "address": "127.0.0.1",
              "port_value": 8001
            }
          }
        ],
        "circuit_breakers": {
          "thresholds": [
            {
              "priority": "HIGH",
              "max_connections": 100000,
              "max_pending_requests": 100000,
              "max_requests": 60000000,
              "max_retries": 50
            },
            {
              "max_connections": 100000,
              "max_pending_requests": 100000,
              "max_requests": 60000000,
              "max_retries": 50
            }
          ]
        },
        "http2_protocol_options": {}
      },
      {
        "name": "service_stats",
        "type": "LOGICAL_DNS",
        "connect_timeout": "0.250s",
        "hosts": [
          {
            "socket_address": {
              "address": "127.0.0.1",
              "port_value": 9001
            }
          }
        ]
      }
    ]
  },
  "dynamic_resources": {
    "lds_config": {
      "api_config_source": {
        "api_type": "GRPC",
        "grpc_services": [
          {
            "envoy_grpc": {
              "cluster_name": "contour"
            }
          }
        ]
      }
    },
    "cds_config": {
      "api_config_source": {
        "api_type": "GRPC",
        "grpc_services": [
          {
            "envoy_grpc": {
              "cluster_name": "contour"
            }
          }
        ]
      }
    }
  },
  "admin": {
    "access_log_path": "/dev/null",
    "address": {
      "socket_address": {
        "address": "127.0.0.1",
        "port_value": 9001
      }
    }
  }
}`,
		},
		"--statsd-enabled": {
			config: BootstrapConfig{
				StatsdEnabled: true,
			},
			want: `{
  "static_resources": {
    "listeners": [
      {
        "address": {
          "socket_address": {
            "address": "0.0.0.0",
            "port_value": 8002
          }
        },
        "filter_chains": [
          {
            "filters": [
              {
                "name": "envoy.http_connection_manager",
                "config": {
                  "codec_type": "AUTO",
                  "http_filters": [
                    {
                      "config": {
                        "headers": [
                          {
                            "exact_match": "/healthz",
                            "name": ":path"
                          }
                        ],
                        "pass_through_mode": "false"
                      },
                      "name": "envoy.health_check"
                    },
                    {
                      "name": "envoy.router"
                    }
                  ],
                  "route_config": {
                    "virtual_hosts": {
                      "domains": [
                        "*"
                      ],
                      "name": "backend",
                      "routes": [
                        {
                          "match": {
                            "prefix": "/stats"
                          },
                          "route": {
                            "cluster": "service_stats"
                          }
                        }
                      ]
                    }
                  },
                  "stat_prefix": "stats"
                }
              }
            ]
          }
        ]
      }
    ],
    "clusters": [
      {
        "name": "contour",
        "type": "STRICT_DNS",
        "connect_timeout": "5s",
        "hosts": [
          {
            "socket_address": {
              "address": "127.0.0.1",
              "port_value": 8001
            }
          }
        ],
        "circuit_breakers": {
          "thresholds": [
            {
              "priority": "HIGH",
              "max_connections": 100000,
              "max_pending_requests": 100000,
              "max_requests": 60000000,
              "max_retries": 50
            },
            {
              "max_connections": 100000,
              "max_pending_requests": 100000,
              "max_requests": 60000000,
              "max_retries": 50
            }
          ]
        },
        "http2_protocol_options": {}
      },
      {
        "name": "service_stats",
        "type": "LOGICAL_DNS",
        "connect_timeout": "0.250s",
        "hosts": [
          {
            "socket_address": {
              "address": "127.0.0.1",
              "port_value": 9001
            }
          }
        ]
      }
    ]
  },
  "dynamic_resources": {
    "lds_config": {
      "api_config_source": {
        "api_type": "GRPC",
        "grpc_services": [
          {
            "envoy_grpc": {
              "cluster_name": "contour"
            }
          }
        ]
      }
    },
    "cds_config": {
      "api_config_source": {
        "api_type": "GRPC",
        "grpc_services": [
          {
            "envoy_grpc": {
              "cluster_name": "contour"
            }
          }
        ]
      }
    }
  },
  "stats_sinks": [
    {
      "name": "envoy.statsd",
      "config": {
        "address": {
          "socket_address": {
            "address": "127.0.0.1",
            "port_value": "9125",
            "protocol": "UDP"
          }
        }
      }
    }
  ],
  "admin": {
    "access_log_path": "/dev/null",
    "address": {
      "socket_address": {
        "address": "127.0.0.1",
        "port_value": 9001
      }
    }
  }
}`,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := Bootstrap(&tc.config)
			want := new(bootstrap.Bootstrap)
			unmarshal(t, tc.want, want)
			if diff := cmp.Diff(got, want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func unmarshal(t *testing.T, data string, pb proto.Message) {
	err := jsonpb.UnmarshalString(data, pb)
	checkErr(t, err)
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}