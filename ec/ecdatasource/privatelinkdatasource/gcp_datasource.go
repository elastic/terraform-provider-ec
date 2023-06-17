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

func GcpDataSource() datasource.DataSource {
	return &gcpDataSource{
		privateLinkDataSource: privateLinkDataSource[v0GcpModel]{
			csp:             "gcp",
			privateLinkName: "private_service_connect",
		},
	}
}

type gcpDataSource struct {
	privateLinkDataSource[v0GcpModel]
}

func (d *gcpDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about the GCP Private Service Connect configuration for a given region. Further documentation on how to establish a PrivateLink connection can be found in the ESS [documentation](https://www.elastic.co/guide/en/cloud/current/ec-traffic-filtering-psc.html).",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "Region to retrieve the Prive Link configuration for.",
				Required:    true,
			},

			// Computed
			"service_attachment_uri": schema.StringAttribute{
				Description: "The service attachment URI to attach the PSC endpoint to.",
				Computed:    true,
			},
			"domain_name": schema.StringAttribute{
				Description: "The domain name to point towards the PSC endpoint.",
				Computed:    true,
			},
		},
	}
}

type v0GcpModel struct {
	RegionField          string  `tfsdk:"region"`
	ServiceAttachmentUri *string `tfsdk:"service_attachment_uri"`
	DomainName           *string `tfsdk:"domain_name"`
}

func (m v0GcpModel) Region() string {
	return m.RegionField
}
