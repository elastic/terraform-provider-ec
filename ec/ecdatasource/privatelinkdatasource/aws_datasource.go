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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AwsDataSource() datasource.DataSource {
	return &awsDataSource{
		privateLinkDataSource: privateLinkDataSource[v0AwsModel]{
			csp:             "aws",
			privateLinkName: "privatelink",
		},
	}
}

type awsDataSource struct {
	privateLinkDataSource[v0AwsModel]
}

func (d *awsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about the AWS Private Link configuration for a given region. Further documentation on how to establish a PrivateLink connection can be found in the ESS [documentation](https://www.elastic.co/guide/en/cloud/current/ec-traffic-filtering-vpc.html).",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "Region to retrieve the Private Link configuration for.",
				Required:    true,
			},

			// Computed
			"vpc_service_name": schema.StringAttribute{
				Description: "The VPC service name used to connect to the region.",
				Computed:    true,
			},
			"domain_name": schema.StringAttribute{
				Description: "The domain name to used in when configuring a private hosted zone in the VPCE connection.",
				Computed:    true,
			},
			"zone_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the availability zones hosting the VPC endpoints.",
				Computed:    true,
			},
		},
	}
}

type v0AwsModel struct {
	RegionField    string    `tfsdk:"region"`
	VpcServiceName *string   `tfsdk:"vpc_service_name"`
	DomainName     *string   `tfsdk:"domain_name"`
	ZoneIDs        *[]string `tfsdk:"zone_ids"`
}

func (m v0AwsModel) Region() string {
	return m.RegionField
}
