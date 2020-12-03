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

package deploymentsdatasource

import (
	"context"
	"strconv"
	"time"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DataSource returns the ec_deployments data source schema.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: read,

		Schema: newSchema(),

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.API)

	query, err := expandFilters(d)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := deploymentapi.Search(deploymentapi.SearchParams{
		API:     client,
		Request: query,
	})
	if err != nil {
		return diag.FromErr(multierror.NewPrefixed("failed searching deployments", err))
	}

	if err := modelToState(d, res); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func modelToState(d *schema.ResourceData, res *models.DeploymentsSearchResponse) error {
	if d.Id() == "" {
		if b, _ := res.MarshalBinary(); len(b) > 0 {
			d.SetId(strconv.Itoa(schema.HashString(string(b))))
		}
	}

	if err := d.Set("return_count", res.ReturnCount); err != nil {
		return err
	}

	var result = make([]interface{}, 0, len(res.Deployments))
	for _, deployment := range res.Deployments {
		var m = make(map[string]interface{})

		m["deployment_id"] = *deployment.ID

		if len(deployment.Resources.Elasticsearch) > 0 {
			m["elasticsearch_resource_id"] = *deployment.Resources.Elasticsearch[0].ID
		}

		if len(deployment.Resources.Kibana) > 0 {
			m["kibana_resource_id"] = *deployment.Resources.Kibana[0].ID
		}

		if len(deployment.Resources.Apm) > 0 {
			m["apm_resource_id"] = *deployment.Resources.Apm[0].ID
		}

		if len(deployment.Resources.EnterpriseSearch) > 0 {
			m["enterprise_search_resource_id"] = *deployment.Resources.EnterpriseSearch[0].ID
		}

		result = append(result, m)

		if len(result) > 0 {
			if err := d.Set("deployments", result); err != nil {
				return err
			}
		}
	}

	return nil
}
