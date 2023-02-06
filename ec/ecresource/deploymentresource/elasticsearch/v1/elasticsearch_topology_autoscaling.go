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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopologyAutoscalingTF struct {
	MaxSizeResource    types.String `tfsdk:"max_size_resource"`
	MaxSize            types.String `tfsdk:"max_size"`
	MinSizeResource    types.String `tfsdk:"min_size_resource"`
	MinSize            types.String `tfsdk:"min_size"`
	PolicyOverrideJson types.String `tfsdk:"policy_override_json"`
}

type ElasticsearchTopologyAutoscaling struct {
	MaxSizeResource    *string `tfsdk:"max_size_resource"`
	MaxSize            *string `tfsdk:"max_size"`
	MinSizeResource    *string `tfsdk:"min_size_resource"`
	MinSize            *string `tfsdk:"min_size"`
	PolicyOverrideJson *string `tfsdk:"policy_override_json"`
}

type ElasticsearchTopologyAutoscalings []ElasticsearchTopologyAutoscaling
