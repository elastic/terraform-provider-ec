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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newSampleTrafficFilter(id string) modelV0 {
	return modelV0{
		ID:               types.String{Value: id},
		Name:             types.String{Value: "my traffic filter"},
		Type:             types.String{Value: "ip"},
		IncludeByDefault: types.Bool{Value: false},
		Region:           types.String{Value: "us-east-1"},
		Description:      types.String{Null: true},
		Rule: types.Set{
			ElemType: trafficFilterRuleElemType(),
			Elems: []attr.Value{
				newSampleTrafficFilterRule("1.1.1.1", "", "", "", ""),
				newSampleTrafficFilterRule("0.0.0.0/0", "", "", "", ""),
			},
		},
	}
}

func newSampleTrafficFilterRule(source string, description string, azureEndpointName string, azureEndpointGUID string, id string) types.Object {
	return types.Object{
		AttrTypes: trafficFilterRuleAttrTypes(),
		Attrs: map[string]attr.Value{
			"source":              types.String{Value: source, Null: source == ""},
			"description":         types.String{Value: description, Null: description == ""},
			"azure_endpoint_name": types.String{Value: azureEndpointName, Null: azureEndpointName == ""},
			"azure_endpoint_guid": types.String{Value: azureEndpointGUID, Null: azureEndpointGUID == ""},
			"id":                  types.String{Value: id},
		},
	}
}
