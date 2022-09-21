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

func elasticsearchResourceInfoSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Instance configuration of the Elasticsearch resource kind.",
		Computed:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"autoscale": {
				Type:        types.StringType,
				Description: "Whether or not Elasticsearch autoscaling is enabled.",
				Computed:    true,
			},
			"healthy": {
				Type:        types.BoolType,
				Description: "Resource kind health status.",
				Computed:    true,
			},
			"cloud_id": {
				Type:                types.StringType,
				Description:         "The encoded Elasticsearch credentials to use in Beats or Logstash.",
				MarkdownDescription: "The encoded Elasticsearch credentials to use in Beats or Logstash. See [Configure Beats and Logstash with Cloud ID](https://www.elastic.co/guide/en/cloud/current/ec-cloud-id.html) for more information.",
				Computed:            true,
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
			"topology": elasticsearchTopologySchema(),
		}),
	}
}

func elasticsearchResourceInfoAttrTypes() map[string]attr.Type {
	return elasticsearchResourceInfoSchema().Attributes.Type().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func elasticsearchTopologySchema() tfsdk.Attribute {
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
			"node_type_data": {
				Type:        types.BoolType,
				Description: "Defines whether this node can hold data (<7.10.0).",
				Computed:    true,
			},
			"node_type_master": {
				Type:        types.BoolType,
				Description: " Defines whether this node can be elected master (<7.10.0).",
				Computed:    true,
			},
			"node_type_ingest": {
				Type:        types.BoolType,
				Description: "Defines whether this node can run an ingest pipeline (<7.10.0).",
				Computed:    true,
			},
			"node_type_ml": {
				Type:        types.BoolType,
				Description: "Defines whether this node can run ML jobs (<7.10.0).",
				Computed:    true,
			},
			"node_roles": {
				Type:        types.SetType{ElemType: types.StringType},
				Description: "Defines the list of Elasticsearch node roles assigned to the topology element (>=7.10.0).",
				Computed:    true,
			},
			"autoscaling": elasticsearchAutoscalingSchema(),
		}),
	}
}

func elasticsearchTopologyAttrTypes() map[string]attr.Type {
	return elasticsearchTopologySchema().Attributes.Type().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func elasticsearchAutoscalingSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources.",
		Computed:    true,
		Validators:  []tfsdk.AttributeValidator{listvalidator.SizeAtMost(1)},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"max_size_resource": {
				Type:        types.StringType,
				Description: "Maximum resource type for the maximum autoscaling setting.",
				Computed:    true,
			},
			"max_size": {
				Type:        types.StringType,
				Description: "Maximum size value for the maximum autoscaling setting.",
				Computed:    true,
			},
			"min_size_resource": {
				Type:        types.StringType,
				Description: "Minimum resource type for the minimum autoscaling setting.",
				Computed:    true,
			},
			"min_size": {
				Type:        types.StringType,
				Description: "Minimum size value for the minimum autoscaling setting.",
				Computed:    true,
			},
			"policy_override_json": {
				Type:        types.StringType,
				Description: "Computed policy overrides set directly via the API or other clients.",
				Computed:    true,
			},
		}),
	}
}

func elasticsearchAutoscalingListType() attr.Type {
	return elasticsearchAutoscalingSchema().Attributes.Type()
}

func elasticsearchAutoscalingAttrTypes() map[string]attr.Type {
	return elasticsearchAutoscalingListType().(types.ListType).ElemType.(types.ObjectType).AttrTypes

}

type elasticsearchResourceInfoModelV0 struct {
	Autoscale     types.String `tfsdk:"autoscale"`
	Healthy       types.Bool   `tfsdk:"healthy"`
	CloudID       types.String `tfsdk:"cloud_id"`
	HttpEndpoint  types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint types.String `tfsdk:"https_endpoint"`
	RefID         types.String `tfsdk:"ref_id"`
	ResourceID    types.String `tfsdk:"resource_id"`
	Status        types.String `tfsdk:"status"`
	Version       types.String `tfsdk:"version"`
	Topology      types.List   `tfsdk:"topology"` //< elasticsearchTopologyModelV0
}

type elasticsearchTopologyModelV0 struct {
	InstanceConfigurationID types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
	NodeTypeData            types.Bool   `tfsdk:"node_type_data"`
	NodeTypeMaster          types.Bool   `tfsdk:"node_type_master"`
	NodeTypeIngest          types.Bool   `tfsdk:"node_type_ingest"`
	NodeTypeMl              types.Bool   `tfsdk:"node_type_ml"`
	NodeRoles               types.Set    `tfsdk:"node_roles"`
	Autoscaling             types.List   `tfsdk:"autoscaling"` //< elasticsearchAutoscalingModel
}

type elasticsearchAutoscalingModel struct {
	MaxSizeResource    types.String `tfsdk:"max_size_resource"`
	MaxSize            types.String `tfsdk:"max_size"`
	MinSizeResource    types.String `tfsdk:"min_size_resource"`
	MinSize            types.String `tfsdk:"min_size"`
	PolicyOverrideJson types.String `tfsdk:"policy_override_json"`
}
