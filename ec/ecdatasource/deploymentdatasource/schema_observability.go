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

package deploymentdatasource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func observabilitySettingsSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Observability settings. Information about logs and metrics shipped to a dedicated deployment.",
		Computed:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"deployment_id": {
				Type:        types.StringType,
				Description: "Destination deployment ID for the shipped logs and monitoring metrics.",
				Computed:    true,
			},
			"ref_id": {
				Type:        types.StringType,
				Description: "Elasticsearch resource kind ref_id of the destination deployment.",
				Computed:    true,
			},
			"logs": {
				Type:        types.BoolType,
				Description: "Defines whether logs are enabled or disabled.",
				Computed:    true,
			},
			"metrics": {
				Type:        types.BoolType,
				Description: "Defines whether metrics are enabled or disabled.",
				Computed:    true,
			},
		}),
	}
}

func observabilitySettingsAttrTypes() map[string]attr.Type {
	return observabilitySettingsSchema().Attributes.Type().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

type observabilitySettingsModel struct {
	DeploymentID types.String `tfsdk:"deployment_id"`
	RefID        types.String `tfsdk:"ref_id"`
	Logs         types.Bool   `tfsdk:"logs"`
	Metrics      types.Bool   `tfsdk:"metrics"`
}
