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
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func Test_AzureDataSource_ReadContext(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		diag     diag.Diagnostics
		endpoint *schema.ResourceData
	}{
		{
			name:   "invalid region returns unknown regino error",
			region: "unknown",
			diag:   diag.FromErr(fmt.Errorf("%w: unknown", errUnknownRegion)),
			endpoint: util.NewResourceData(t, util.ResDataParams{
				ID: "myID",
				State: map[string]interface{}{
					"id":     "myID",
					"region": "unknown",
				},
				Schema: newAzureSchema(),
			}),
		},
		{
			name:   "valid region returns endpoint",
			region: "uksouth",
			endpoint: util.NewResourceData(t, util.ResDataParams{
				ID: "myID",
				State: map[string]interface{}{
					"id":            "myID",
					"region":        "uksouth",
					"service_alias": "uksouth-prod-007-privatelink-service.98758729-06f7-438d-baaa-0cb63e737cdf.uksouth.azure.privatelinkservice",
					"domain_name":   "privatelink.uksouth.azure.elastic-cloud.com",
				},
				Schema: newAzureSchema(),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentsSchemaArg := schema.TestResourceDataRaw(t, newAzureSchema(), nil)
			deploymentsSchemaArg.SetId("myID")
			_ = deploymentsSchemaArg.Set("region", tt.region)

			source := AzureDataSource()

			d := source.ReadContext(context.Background(), deploymentsSchemaArg, nil)
			if tt.diag != nil {
				assert.Equal(t, d, tt.diag)
			} else {
				assert.Nil(t, d)
			}

			assert.Equal(t, tt.endpoint.State().Attributes, deploymentsSchemaArg.State().Attributes)
		})
	}
}
