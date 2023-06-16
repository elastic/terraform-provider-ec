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
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func observabilitySettingsSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Observability settings. Information about logs and metrics shipped to a dedicated deployment.",
		Computed:    true,
		Validators:  []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"deployment_id": schema.StringAttribute{
					Description: "Destination deployment ID for the shipped logs and monitoring metrics.",
					Computed:    true,
				},
				"ref_id": schema.StringAttribute{
					Description: "Elasticsearch resource kind ref_id of the destination deployment.",
					Computed:    true,
				},
				"logs": schema.BoolAttribute{
					Description: "Defines whether logs are shipped to the destination deployment.",
					Computed:    true,
				},
				"metrics": schema.BoolAttribute{
					Description: "Defines whether metrics are shipped to the destination deployment.",
					Computed:    true,
				},
			},
		},
	}
}

func observabilitySettingsAttrTypes() map[string]attr.Type {
	return observabilitySettingsSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

type observabilitySettingsModel struct {
	DeploymentID types.String `tfsdk:"deployment_id"`
	RefID        types.String `tfsdk:"ref_id"`
	Logs         types.Bool   `tfsdk:"logs"`
	Metrics      types.Bool   `tfsdk:"metrics"`
}
