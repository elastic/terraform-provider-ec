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

package privatelinkdatasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func AzureDataSource() datasource.DataSource {
	return &azureDataSource{
		privateLinkDataSource: privateLinkDataSource[v0AzureModel]{
			csp:             "azure",
			privateLinkName: "privatelink",
		},
	}
}

type azureDataSource struct {
	privateLinkDataSource[v0AzureModel]
}

func (d *azureDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about the Azure Private Link configuration for a given region. Further documentation on how to establish a PrivateLink connection can be found in the ESS [documentation](https://www.elastic.co/guide/en/cloud/current/ec-traffic-filtering-vnet.html).",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "Region to retrieve the Private Link configuration for.",
				Required:    true,
			},

			// Computed
			"service_alias": schema.StringAttribute{
				Description: "The service alias to establish a connection to.",
				Computed:    true,
			},
			"domain_name": schema.StringAttribute{
				Description: "The domain name to used in when configuring a private hosted zone in the VNet connection.",
				Computed:    true,
			},
		},
	}
}

type v0AzureModel struct {
	RegionField  string  `tfsdk:"region"`
	ServiceAlias *string `tfsdk:"service_alias"`
	DomainName   *string `tfsdk:"domain_name"`
}

func (m v0AzureModel) Region() string {
	return m.RegionField
}
