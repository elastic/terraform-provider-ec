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

package trafficfilterresource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func newSampleTrafficFilter(t *testing.T, id string) modelV0 {
	return modelV0{
		ID:               types.StringValue(id),
		Name:             types.StringValue("my traffic filter"),
		Type:             types.StringValue("ip"),
		IncludeByDefault: types.BoolValue(false),
		Region:           types.StringValue("us-east-1"),
		Description:      types.StringNull(),
		Rule: func() types.Set {
			res, diags := types.SetValue(
				trafficFilterRuleElemType(),
				[]attr.Value{
					newSampleTrafficFilterRule(t, "1.1.1.1", "", "", "", ""),
					newSampleTrafficFilterRule(t, "0.0.0.0/0", "", "", "", ""),
				},
			)
			assert.Nil(t, diags)
			return res
		}(),
	}
}

func newSampleTrafficFilterRule(t *testing.T, source string, description string, azureEndpointName string, azureEndpointGUID string, id string) types.Object {
	res, diags := types.ObjectValue(
		trafficFilterRuleAttrTypes(),
		map[string]attr.Value{
			"source": func() attr.Value {
				if source == "" {
					return types.StringNull()
				}
				return types.StringValue(source)
			}(),
			"description": func() attr.Value {
				if description == "" {
					return types.StringNull()
				}
				return types.StringValue(description)
			}(),
			"azure_endpoint_name": func() attr.Value {
				if azureEndpointName == "" {
					return types.StringNull()
				}
				return types.StringValue(azureEndpointName)
			}(),
			"azure_endpoint_guid": func() attr.Value {
				if azureEndpointGUID == "" {
					return types.StringNull()
				}
				return types.StringValue(azureEndpointGUID)
			}(),
			"id": types.StringValue(id),
		},
	)
	assert.Nil(t, diags)
	return res
}
