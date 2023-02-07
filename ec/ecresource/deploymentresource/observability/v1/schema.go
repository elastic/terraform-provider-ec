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

package v1

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func ObservabilitySchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Optional observability settings. Ship logs and metrics to a dedicated deployment.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"deployment_id": schema.StringAttribute{
					Required: true,
				},
				"ref_id": schema.StringAttribute{
					Computed: true,
					Optional: true,
				},
				"logs": schema.BoolAttribute{
					Optional: true,
					Computed: true,
				},
				"metrics": schema.BoolAttribute{
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}
