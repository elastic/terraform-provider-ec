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

package deploymenttemplates

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of available deployment templates.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "Region to select. For Elastic Cloud Enterprise (ECE) installations, use `ece-region`.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Filters for a deployment template with this id.",
				Optional:    true,
			},
			"stack_version": schema.StringAttribute{
				Description: "Filters for deployment templates compatible with this stack version.",
				Optional:    true,
			},
			"show_deprecated": schema.BoolAttribute{
				Description: "Enable to also show deprecated deployment templates. (Set to false by default.)",
				Optional:    true,
			},
			"templates": deploymentTemplatesListSchema(),
		},
	}
}

func deploymentTemplatesListSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "List of available deployment templates.",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description: "The id of the deployment template.",
					Computed:    true,
				},
				"name": schema.StringAttribute{
					Description: "The name of the deployment template.",
					Computed:    true,
				},
				"description": schema.StringAttribute{
					Description: "The description of the deployment template.",
					Computed:    true,
				},
				"min_stack_version": schema.StringAttribute{
					Description: "The minimum stack version that can used with this deployment template.",
					Computed:    true,
				},
				"deprecated": schema.BoolAttribute{
					Description: "Outdated templates are marked as deprecated, but can still be used.",
					Computed:    true,
				},
				"elasticsearch":       elasticsearchSchema(),
				"kibana":              statelessSchema(),
				"enterprise_search":   statelessSchema(),
				"apm":                 statelessSchema(),
				"integrations_server": statelessSchema(),
			},
		},
	}
}

func elasticsearchSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Defines the default configuration for Elasticsearch.",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"hot":          topologySchema(),
			"coordinating": topologySchema(),
			"master":       topologySchema(),
			"warm":         topologySchema(),
			"cold":         topologySchema(),
			"frozen":       topologySchema(),
			"ml":           topologySchema(),
		},
	}
}

func topologySchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Defines the default configuration for each topology.",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"instance_configuration_id": schema.StringAttribute{
				Computed: true,
			},
			"instance_configuration_version": schema.NumberAttribute{
				Computed: true,
			},
			"default_size": schema.StringAttribute{
				Computed: true,
			},
			"available_sizes": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"size_resource": schema.StringAttribute{
				Computed: true,
			},
			"autoscaling": autoscalingSchema(),
		},
	}
}

func autoscalingSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"max_size_resource": schema.StringAttribute{
				Computed: true,
			},
			"max_size": schema.StringAttribute{
				Computed: true,
			},
			"min_size_resource": schema.StringAttribute{
				Computed: true,
			},
			"min_size": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func statelessSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Defines the default configuration for a stateless application (Kibana, Enterprise Search, APM or Integrations Server).",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"instance_configuration_id": schema.StringAttribute{
				Computed: true,
			},
			"instance_configuration_version": schema.NumberAttribute{
				Computed: true,
			},
			"default_size": schema.StringAttribute{
				Computed: true,
			},
			"available_sizes": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"size_resource": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}
