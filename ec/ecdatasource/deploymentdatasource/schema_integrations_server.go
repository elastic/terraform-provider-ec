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

func integrationsServerResourceInfoSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Instance configuration of the Integrations Server type.",
		Computed:    true,
		Validators:  []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"elasticsearch_cluster_ref_id": schema.StringAttribute{
					Description: "A locally-unique friendly alias for an Elasticsearch resource in this deployment.",
					Computed:    true,
				},
				"healthy": schema.BoolAttribute{
					Description: "Resource kind health status.",
					Computed:    true,
				},
				"http_endpoint": schema.StringAttribute{
					Description: "HTTP endpoint for the resource kind.",
					Computed:    true,
				},
				"https_endpoint": schema.StringAttribute{
					Description: "HTTPS endpoint for the resource kind.",
					Computed:    true,
				},
				"ref_id": schema.StringAttribute{
					Description: "A locally-unique friendly alias for this Integrations Server resource.",
					Computed:    true,
				},
				"resource_id": schema.StringAttribute{
					Description: "The resource unique identifier.",
					Computed:    true,
				},
				"status": schema.StringAttribute{
					Description: "Resource kind status (for example, \"started\", \"stopped\", etc).",
					Computed:    true,
				},
				"version": schema.StringAttribute{
					Description: "Elastic stack version.",
					Computed:    true,
				},
				"topology": integrationsServerTopologySchema(),
				"fleet_endpoint": schema.StringAttribute{
					Description: "HTTPS endpoint for the fleet.",
					Computed:    true,
				},
			},
		},
	}
}

func integrationsServerResourceInfoAttrTypes() map[string]attr.Type {
	return integrationsServerResourceInfoSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func integrationsServerTopologySchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Node topology element definition.",
		Computed:    true,
		Validators:  []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"instance_configuration_id": schema.StringAttribute{
					Description: "Controls the allocation of this topology element as well as allowed sizes and node_types. It needs to match the ID of an existing instance configuration.",
					Computed:    true,
				},
				"size": schema.StringAttribute{
					Description: `Amount of "size_resource" in Gigabytes. For example "4g".`,
					Computed:    true,
				},
				"size_resource": schema.StringAttribute{
					Description: "Type of resource (\"memory\" or \"storage\")",
					Computed:    true,
				},
				"zone_count": schema.Int64Attribute{
					Description: "Number of zones in which nodes will be placed.",
					Computed:    true,
				},
			},
		},
	}
}

func integrationsServerTopologyAttrTypes() map[string]attr.Type {
	return integrationsServerTopologySchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

type integrationsServerResourceInfoModelV0 struct {
	ElasticsearchClusterRefID types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	Healthy                   types.Bool   `tfsdk:"healthy"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	RefID                     types.String `tfsdk:"ref_id"`
	ResourceID                types.String `tfsdk:"resource_id"`
	Status                    types.String `tfsdk:"status"`
	Version                   types.String `tfsdk:"version"`
	Topology                  types.List   `tfsdk:"topology"` //< integrationsServerTopologyModelV0
	FleetEndpoint             types.String `tfsdk:"fleet_endpoint"`
}

type integrationsServerTopologyModelV0 struct {
	InstanceConfigurationID types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
}
