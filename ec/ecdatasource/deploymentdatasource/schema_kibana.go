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

func kibanaResourceInfoSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Instance configuration of the Kibana type.",
		Computed:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"elasticsearch_cluster_ref_id": {
				Type:        types.StringType,
				Description: "The user-specified ID of the Elasticsearch cluster to which this resource kind will link.",
				Computed:    true,
			},
			"healthy": {
				Type:        types.BoolType,
				Description: "Resource kind health status.",
				Computed:    true,
			},
			"http_endpoint": {
				Type:        types.StringType,
				Description: "HTTP endpoint for the resource kind.",
				Computed:    true,
			},
			"https_endpoint": {
				Type:        types.StringType,
				Description: "HTTPS endpoint for the resource kind.",
				Computed:    true,
			},
			"ref_id": {
				Type:        types.StringType,
				Description: "User specified ref_id for the resource kind.",
				Computed:    true,
			},
			"resource_id": {
				Type:        types.StringType,
				Description: "The resource unique identifier.",
				Computed:    true,
			},
			"status": {
				Type:        types.StringType,
				Description: "Resource kind status (for example, \"started\", \"stopped\", etc).",
				Computed:    true,
			},
			"version": {
				Type:        types.StringType,
				Description: "Elastic stack version.",
				Computed:    true,
			},
			"topology": kibanaTopologySchema(),
		}),
	}
}

func kibanaResourceInfoAttrTypes() map[string]attr.Type {
	return kibanaResourceInfoSchema().Attributes.Type().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func kibanaTopologySchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Node topology element definition.",
		Computed:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"instance_configuration_id": {
				Type:        types.StringType,
				Description: "Controls the allocation of this topology element as well as allowed sizes and node_types. It needs to match the ID of an existing instance configuration.",
				Computed:    true,
			},
			"size": {
				Type:        types.StringType,
				Description: "Amount of resource per topology element in the \"g\" notation.",
				Computed:    true,
			},
			"size_resource": {
				Type:        types.StringType,
				Description: "Type of resource (\"memory\" or \"storage\")",
				Computed:    true,
			},
			"zone_count": {
				Type:        types.Int64Type,
				Description: "Number of zones in which nodes will be placed.",
				Computed:    true,
			},
		}),
	}
}

func kibanaTopologyAttrTypes() map[string]attr.Type {
	return kibanaTopologySchema().Attributes.Type().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

type kibanaResourceInfoModelV0 struct {
	ElasticsearchClusterRefID types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	Healthy                   types.Bool   `tfsdk:"healthy"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	RefID                     types.String `tfsdk:"ref_id"`
	ResourceID                types.String `tfsdk:"resource_id"`
	Status                    types.String `tfsdk:"status"`
	Version                   types.String `tfsdk:"version"`
	Topology                  types.List   `tfsdk:"topology"` //< kibanaTopologyModelV0
}

type kibanaTopologyModelV0 struct {
	InstanceConfigurationID types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
}
