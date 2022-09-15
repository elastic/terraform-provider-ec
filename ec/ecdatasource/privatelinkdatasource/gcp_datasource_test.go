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

func Test_GcpDataSource_ReadContext(t *testing.T) {
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
				Schema: newGcpSchema(),
			}),
		},
		{
			name:   "valid region returns endpoint",
			region: "us-central1",
			endpoint: util.NewResourceData(t, util.ResDataParams{
				ID: "myID",
				State: map[string]interface{}{
					"id":                     "myID",
					"region":                 "us-central1",
					"service_attachment_uri": "projects/cloud-production-168820/regions/us-central1/serviceAttachments/proxy-psc-production-us-central1-v1-attachment",
					"domain_name":            "psc.us-central1.gcp.cloud.es.io",
				},
				Schema: newGcpSchema(),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentsSchemaArg := schema.TestResourceDataRaw(t, newGcpSchema(), nil)
			deploymentsSchemaArg.SetId("myID")
			_ = deploymentsSchemaArg.Set("region", tt.region)

			source := GcpDataSource()

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
