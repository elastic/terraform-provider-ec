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

func elasticsearchResourceInfoSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Instance configuration of the Elasticsearch Elasticsearch resource.",
		Computed:    true,
		Validators:  []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"autoscale": schema.StringAttribute{
					Description: "Whether or not Elasticsearch autoscaling is enabled.",
					Computed:    true,
				},
				"healthy": schema.BoolAttribute{
					Description: "Elasticsearch resource health status.",
					Computed:    true,
				},
				"cloud_id": schema.StringAttribute{
					Description:         "The cloud ID, an encoded string that provides other Elastic services with the necessary information to connect to this Elasticsearch and Kibana.",
					MarkdownDescription: "The cloud ID, an encoded string that provides other Elastic services with the necessary information to connect to this Elasticsearch and Kibana. See [Configure Beats and Logstash with Cloud ID](https://www.elastic.co/guide/en/cloud/current/ec-cloud-id.html) for more information.",
					Computed:            true,
				},
				"http_endpoint": schema.StringAttribute{
					Description: "HTTP endpoint for the Elasticsearch resource.",
					Computed:    true,
				},
				"https_endpoint": schema.StringAttribute{
					Description: "HTTPS endpoint for the Elasticsearch resource.",
					Computed:    true,
				},
				"ref_id": schema.StringAttribute{
					Description: "A locally-unique friendly alias for this Elasticsearch cluster.",
					Computed:    true,
				},
				"resource_id": schema.StringAttribute{
					Description: "The resource unique identifier.",
					Computed:    true,
				},
				"status": schema.StringAttribute{
					Description: "Elasticsearch resource status (for example, \"started\", \"stopped\", etc).",
					Computed:    true,
				},
				"version": schema.StringAttribute{
					Description: "Elastic stack version.",
					Computed:    true,
				},
				"topology": elasticsearchTopologySchema(),
			},
		},
	}
}

func elasticsearchResourceInfoAttrTypes() map[string]attr.Type {
	return elasticsearchResourceInfoSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func elasticsearchTopologySchema() schema.Attribute {
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
					Description: `Amount of "size_resource" per topology element in Gigabytes. For example "4g".`,
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
				"node_type_data": schema.BoolAttribute{
					Description: "Defines whether this node can hold data (<8.0).",
					Computed:    true,
				},
				"node_type_master": schema.BoolAttribute{
					Description: "Defines whether this node can be elected master (<8.0).",
					Computed:    true,
				},
				"node_type_ingest": schema.BoolAttribute{
					Description: "Defines whether this node can run an ingest pipeline (<8.0).",
					Computed:    true,
				},
				"node_type_ml": schema.BoolAttribute{
					Description: "Defines whether this node can run ML jobs (<8.0).",
					Computed:    true,
				},
				"node_roles": schema.SetAttribute{
					ElementType: types.StringType,
					Description: "Defines the list of Elasticsearch node roles assigned to the topology element. This is supported from v7.10, and required from v8.",
					Computed:    true,
				},
				"autoscaling": elasticsearchAutoscalingSchema(),
			},
		},
	}
}

func elasticsearchTopologyAttrTypes() map[string]attr.Type {
	return elasticsearchTopologySchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func elasticsearchAutoscalingSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Description: "Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources.",
		Computed:    true,
		Validators:  []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"max_size_resource": schema.StringAttribute{
					Description: "Resource type used when specifying the maximum size the tier can scale up to.",
					Computed:    true,
				},
				"max_size": schema.StringAttribute{
					Description: "Maximum size the tier can scale up to, e.g \"64g\".",
					Computed:    true,
				},
				"min_size_resource": schema.StringAttribute{
					Description: "Resource type used when specifying the minimum size the tier can scale down to when bidirectional autoscaling is supported.",
					Computed:    true,
				},
				"min_size": schema.StringAttribute{
					Description: "Minimum size the tier can scale down to when bidirectional autoscaling is supported.",
					Computed:    true,
				},
				"policy_override_json": schema.StringAttribute{
					Description: "An arbitrary JSON object overriding the default autoscaling policy. Don't set unless you really know what you are doing.",
					Computed:    true,
				},
			},
		},
	}
}

func elasticsearchAutoscalingListType() attr.Type {
	return elasticsearchAutoscalingSchema().GetType()
}

func elasticsearchAutoscalingElemType() attr.Type {
	return elasticsearchAutoscalingListType().(types.ListType).ElementType()
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
