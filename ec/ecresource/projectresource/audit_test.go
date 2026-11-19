// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package projectresource

import (
	"context"
	"encoding/json"
	"sort"
	"testing"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestExpandMonitoringForCreate_OmitsMonitoringWhenAuditNull(t *testing.T) {
	cases := []struct {
		name string
		in   *monitoring
	}{
		{name: "no monitoring", in: nil},
		{name: "monitoring empty", in: &monitoring{}},
		{name: "logging empty", in: &monitoring{Logging: &logging{}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := marshalCreate(t, expandMonitoringForCreate(tc.in))
			var body map[string]any
			require.NoError(t, json.Unmarshal(raw, &body))
			_, ok := body["monitoring"]
			require.False(t, ok)
		})
	}
}

func TestExpandMonitoringForCreate_SendsAudit(t *testing.T) {
	enabled := false
	raw := marshalCreate(t, expandMonitoringForCreate(&monitoring{
		Logging: &logging{
			Audit: &audit{
				Enabled:         enabled,
				ProjectID:       "proj-1",
				ProjectType:     "observability",
				IgnoreFilterIDs: []string{"audit-ignore-data-reads"},
			},
		},
	}))
	var body struct {
		Monitoring json.RawMessage `json:"monitoring"`
	}
	require.NoError(t, json.Unmarshal(raw, &body))
	require.JSONEq(t, `{
		"logging": {
			"audit": {
				"enabled": false,
				"destination": {
					"project_id": "proj-1",
					"project_type": "observability"
				},
				"ignore_filters": [{"id": "audit-ignore-data-reads"}]
			}
		}
	}`, string(body.Monitoring))
	require.NotContains(t, string(body.Monitoring), "status")
}

func TestExpandMonitoringForCreate_OmitsEmptyIgnoreFilters(t *testing.T) {
	for _, ids := range [][]string{nil, {}} {
		raw := marshalCreate(t, expandMonitoringForCreate(&monitoring{
			Logging: &logging{
				Audit: &audit{
					Enabled:         true,
					ProjectID:       "proj-1",
					ProjectType:     "security",
					IgnoreFilterIDs: ids,
				},
			},
		}))
		var body struct {
			Monitoring json.RawMessage `json:"monitoring"`
		}
		require.NoError(t, json.Unmarshal(raw, &body))
		require.NotContains(t, string(body.Monitoring), "ignore_filters")
		require.Contains(t, string(body.Monitoring), `"enabled":true`)
	}
}

func TestFlattenIgnoreFilterIDs(t *testing.T) {
	require.Nil(t, flattenIgnoreFilterIDs(nil))
	require.Equal(t, []string{}, flattenIgnoreFilterIDs(&[]serverless.IgnoreFilterRef{}))
	filters := []serverless.IgnoreFilterRef{{Id: "audit-ignore-data-reads"}, {Id: "other"}}
	require.Equal(t, []string{"audit-ignore-data-reads", "other"}, flattenIgnoreFilterIDs(&filters))
}

// withAudit is a monitoring block with audit set. ids nil is a null
// ignore_filter_ids set; an empty non-nil slice is an empty set.
func withAudit(enabled bool, ids []string) *monitoring {
	return &monitoring{Logging: &logging{Audit: &audit{
		Enabled:         enabled,
		ProjectID:       "proj-1",
		ProjectType:     "observability",
		IgnoreFilterIDs: ids,
	}}}
}

func TestAuditPatchJSON(t *testing.T) {
	var unset *monitoring

	cases := []struct {
		name     string
		plan     *monitoring
		state    *monitoring
		want     string
		omitKey  string
		contains string
		absent   []string
	}{
		{
			name:  "enabled false still sends destination",
			plan:  withAudit(false, []string{"audit-ignore-data-reads"}),
			state: withAudit(true, []string{"audit-ignore-data-reads"}),
			want: `{
				"name": "proj",
				"monitoring": {
					"logging": {
						"audit": {
							"enabled": false,
							"destination": {"project_id": "proj-1", "project_type": "observability"},
							"ignore_filters": [{"id": "audit-ignore-data-reads"}]
						}
					}
				}
			}`,
			absent: []string{"status"},
		},
		{
			name:  "non-empty filters replaced by null sends ignore_filters null",
			plan:  withAudit(true, nil),
			state: withAudit(true, []string{"audit-ignore-data-reads"}),
			want: `{
				"name": "proj",
				"monitoring": {
					"logging": {
						"audit": {
							"enabled": true,
							"destination": {"project_id": "proj-1", "project_type": "observability"},
							"ignore_filters": null
						}
					}
				}
			}`,
		},
		{
			name:  "non-empty filters replaced by empty set sends ignore_filters null",
			plan:  withAudit(false, []string{}),
			state: withAudit(false, []string{"audit-ignore-data-reads"}),
			want: `{
				"name": "proj",
				"monitoring": {
					"logging": {
						"audit": {
							"enabled": false,
							"destination": {"project_id": "proj-1", "project_type": "observability"},
							"ignore_filters": null
						}
					}
				}
			}`,
		},
		{
			name:     "null filters in plan and state omit ignore_filters",
			plan:     withAudit(true, nil),
			state:    withAudit(true, nil),
			omitKey:  "ignore_filters",
			contains: `"enabled":true`,
		},
		{
			name:    "empty filters in plan and state omit ignore_filters",
			plan:    withAudit(true, []string{}),
			state:   withAudit(true, []string{}),
			omitKey: "ignore_filters",
		},
		{
			name:  "removed audit sends audit null and not monitoring null",
			plan:  unset,
			state: withAudit(true, nil),
			want:  `{"name":"proj","monitoring":{"logging":{"audit":null}}}`,
			absent: []string{
				`"monitoring":null`,
			},
		},
		{
			name:  "audit block removed from a monitoring shell sends audit null",
			plan:  &monitoring{Logging: &logging{}},
			state: withAudit(true, nil),
			want:  `{"name":"proj","monitoring":{"logging":{"audit":null}}}`,
			absent: []string{
				`"monitoring":null`,
			},
		},
		{
			name:  "adding audit sends the audit object and not audit null",
			plan:  withAudit(true, nil),
			state: unset,
			want: `{
				"name": "proj",
				"monitoring": {
					"logging": {
						"audit": {
							"enabled": true,
							"destination": {"project_id": "proj-1", "project_type": "observability"}
						}
					}
				}
			}`,
			absent: []string{`"audit":null`},
		},
		{
			name:    "audit unset in state and config omits monitoring",
			plan:    unset,
			state:   unset,
			omitKey: "monitoring",
		},
		{
			name:    "shell added to unset state omits monitoring",
			plan:    &monitoring{Logging: &logging{}},
			state:   unset,
			omitKey: "monitoring",
		},
		{
			name:    "shell removed from state omits monitoring",
			plan:    unset,
			state:   &monitoring{},
			omitKey: "monitoring",
		},
		{
			name:    "shell in plan and state omits monitoring",
			plan:    &monitoring{},
			state:   &monitoring{Logging: &logging{}},
			omitKey: "monitoring",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := marshalPatch(t, tc.plan, tc.state)
			if tc.want != "" {
				require.JSONEq(t, tc.want, string(raw))
			}
			if tc.omitKey != "" {
				var body map[string]any
				require.NoError(t, json.Unmarshal(raw, &body))
				if tc.omitKey == "monitoring" {
					_, ok := body["monitoring"]
					require.False(t, ok)
				} else {
					require.NotContains(t, string(raw), tc.omitKey)
				}
			}
			if tc.contains != "" {
				require.Contains(t, string(raw), tc.contains)
			}
			for _, absent := range tc.absent {
				require.NotContains(t, string(raw), absent)
			}
		})
	}
}

// TestMonitoringFrom checks the three states a Terraform monitoring value
// maps to: nil for an omitted block, a shell for a block without audit, and
// a configured audit.
func TestMonitoringFrom(t *testing.T) {
	monitoringTypes := resource_elasticsearch_project.MonitoringValue{}.AttributeTypes(context.Background())
	loggingTypes, diags := objectAttrTypes(monitoringTypes, "logging")
	require.False(t, diags.HasError(), diags)
	auditTypes, diags := objectAttrTypes(loggingTypes, "audit")
	require.False(t, diags.HasError(), diags)
	loggingNull := types.ObjectNull(loggingTypes)
	loggingWithoutAudit, diags := types.ObjectValue(loggingTypes, map[string]attr.Value{"audit": types.ObjectNull(auditTypes)})
	require.False(t, diags.HasError(), diags)

	cases := []struct {
		name           string
		monitoringNull bool
		logging        basetypes.ObjectValue
		want           *monitoring
	}{
		{name: "omitted monitoring", monitoringNull: true, logging: loggingNull, want: nil},
		{name: "monitoring without logging is a shell", logging: loggingNull, want: &monitoring{}},
		{name: "logging without audit is a shell", logging: loggingWithoutAudit, want: &monitoring{Logging: &logging{}}},
		{
			name:    "audit is configured",
			logging: loggingWithAudit(t, monitoringTypes),
			want: &monitoring{Logging: &logging{Audit: &audit{
				Enabled:         false,
				ProjectID:       "dest-1",
				ProjectType:     "observability",
				IgnoreFilterIDs: []string{"audit-ignore-data-reads"},
			}}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, diags := monitoringFrom(context.Background(), tc.monitoringNull, tc.logging)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestFlattenMonitoringJSON(t *testing.T) {
	monitoringTypes := resource_elasticsearch_project.MonitoringValue{}.AttributeTypes(context.Background())
	configured := withAudit(true, []string{"audit-ignore-data-reads"})
	emptyFilters := withAudit(true, []string{})
	nullFilters := withAudit(true, nil)
	var monitoringNull *monitoring
	shell := &monitoring{}

	apiWith := func(filters *[]serverless.IgnoreFilterRef) *serverless.Monitoring {
		return &serverless.Monitoring{
			Logging: &serverless.Logging{
				Audit: &serverless.LoggingCategory{
					Enabled: true,
					Destination: serverless.LoggingDestination{
						ProjectId:   "proj-1",
						ProjectType: "observability",
						Status:      "enabled",
					},
					IgnoreFilters: filters,
				},
			},
		}
	}
	ids := []serverless.IgnoreFilterRef{{Id: "other"}, {Id: "audit-ignore-data-reads"}}

	cases := []struct {
		name  string
		prior *monitoring
		api   *serverless.Monitoring
		want  string
	}{
		{
			name:  "stores destination status",
			prior: configured,
			api:   apiWith(&ids),
			want: `{
				"logging": {
					"audit": {
						"enabled": true,
						"destination": {
							"project_id": "proj-1",
							"project_type": "observability",
							"status": "enabled"
						},
						"ignore_filter_ids": ["audit-ignore-data-reads", "other"]
					}
				}
			}`,
		},
		{
			name:  "null monitoring and no API audit stores monitoring null",
			prior: monitoringNull,
			want:  `null`,
		},
		{
			name:  "monitoring shell and no API audit stores the same shell",
			prior: shell,
			want:  `{"logging": null}`,
		},
		{
			name:  "logging shell and no API audit stores the same shell",
			prior: &monitoring{Logging: &logging{}},
			want:  `{"logging": {"audit": null}}`,
		},
		{
			name:  "null monitoring and API audit stores audit with null filters when array is absent",
			prior: monitoringNull,
			api:   apiWith(nil),
			want: `{
				"logging": {
					"audit": {
						"enabled": true,
						"destination": {
							"project_id": "proj-1",
							"project_type": "observability",
							"status": "enabled"
						},
						"ignore_filter_ids": null
					}
				}
			}`,
		},
		{
			name:  "shell and API audit stores audit with null filters when array is empty",
			prior: shell,
			api:   apiWith(&[]serverless.IgnoreFilterRef{}),
			want: `{
				"logging": {
					"audit": {
						"enabled": true,
						"destination": {
							"project_id": "proj-1",
							"project_type": "observability",
							"status": "enabled"
						},
						"ignore_filter_ids": null
					}
				}
			}`,
		},
		{
			name:  "configured audit and no API audit stores monitoring null",
			prior: configured,
			want:  `null`,
		},
		{
			name:  "API ids become the ignore_filter_ids set",
			prior: nullFilters,
			api:   apiWith(&ids),
			want: `{
				"logging": {
					"audit": {
						"enabled": true,
						"destination": {
							"project_id": "proj-1",
							"project_type": "observability",
							"status": "enabled"
						},
						"ignore_filter_ids": ["audit-ignore-data-reads", "other"]
					}
				}
			}`,
		},
		{
			name:  "null filters stay null when the API array is absent",
			prior: nullFilters,
			api:   apiWith(nil),
			want: `{
				"logging": {
					"audit": {
						"enabled": true,
						"destination": {
							"project_id": "proj-1",
							"project_type": "observability",
							"status": "enabled"
						},
						"ignore_filter_ids": null
					}
				}
			}`,
		},
		{
			name:  "null filters stay null when the API array is empty",
			prior: nullFilters,
			api:   apiWith(&[]serverless.IgnoreFilterRef{}),
			want: `{
				"logging": {
					"audit": {
						"enabled": true,
						"destination": {
							"project_id": "proj-1",
							"project_type": "observability",
							"status": "enabled"
						},
						"ignore_filter_ids": null
					}
				}
			}`,
		},
		{
			name:  "empty filters stay an empty set when the API array is absent",
			prior: emptyFilters,
			api:   apiWith(nil),
			want: `{
				"logging": {
					"audit": {
						"enabled": true,
						"destination": {
							"project_id": "proj-1",
							"project_type": "observability",
							"status": "enabled"
						},
						"ignore_filter_ids": []
					}
				}
			}`,
		},
		{
			name:  "empty filters stay an empty set when the API array is empty",
			prior: emptyFilters,
			api:   apiWith(&[]serverless.IgnoreFilterRef{}),
			want: `{
				"logging": {
					"audit": {
						"enabled": true,
						"destination": {
							"project_id": "proj-1",
							"project_type": "observability",
							"status": "enabled"
						},
						"ignore_filter_ids": []
					}
				}
			}`,
		},
		{
			name:  "non-empty filters become null when the API array is absent",
			prior: configured,
			api:   apiWith(nil),
			want: `{
				"logging": {
					"audit": {
						"enabled": true,
						"destination": {
							"project_id": "proj-1",
							"project_type": "observability",
							"status": "enabled"
						},
						"ignore_filter_ids": null
					}
				}
			}`,
		},
		{
			name:  "non-empty filters become null when the API array is empty",
			prior: configured,
			api:   apiWith(&[]serverless.IgnoreFilterRef{}),
			want: `{
				"logging": {
					"audit": {
						"enabled": true,
						"destination": {
							"project_id": "proj-1",
							"project_type": "observability",
							"status": "enabled"
						},
						"ignore_filter_ids": null
					}
				}
			}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj, diags := flattenMonitoring(monitoringTypes, tc.prior, tc.api)
			require.False(t, diags.HasError(), diags)
			require.JSONEq(t, tc.want, monitoringJSON(t, obj))
		})
	}
}

func marshalCreate(t *testing.T, monitoring *serverless.MonitoringCreateRequest) []byte {
	t.Helper()
	raw, err := json.Marshal(serverless.CreateElasticsearchProjectRequest{
		Name:       "proj",
		RegionId:   "aws-us-east-1",
		Monitoring: monitoring,
	})
	require.NoError(t, err)
	return raw
}

func marshalPatch(t *testing.T, plan, state *monitoring) []byte {
	t.Helper()
	name := serverless.ProjectName("proj")
	body := serverless.PatchElasticsearchProjectRequest{Name: &name}
	patch, err := auditPatchFor(plan, state)
	require.NoError(t, err)
	if patch.useRaw() {
		raw, err := mergeMonitoringField(body, patch.raw)
		require.NoError(t, err)
		return raw
	}
	body.Monitoring = patch.typed
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	return raw
}

func monitoringJSON(t *testing.T, obj attr.Value) string {
	t.Helper()
	raw, err := json.Marshal(attrJSON(obj))
	require.NoError(t, err)
	return string(raw)
}

func attrJSON(v attr.Value) any {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return nil
	}
	switch typed := v.(type) {
	case basetypes.StringValue:
		return typed.ValueString()
	case basetypes.BoolValue:
		return typed.ValueBool()
	case basetypes.SetValue:
		ids := make([]string, 0, len(typed.Elements()))
		for _, el := range typed.Elements() {
			ids = append(ids, el.(basetypes.StringValue).ValueString())
		}
		sort.Strings(ids)
		return ids
	case basetypes.ObjectValue:
		out := map[string]any{}
		for key, child := range typed.Attributes() {
			out[key] = attrJSON(child)
		}
		return out
	default:
		return nil
	}
}
