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
	"time"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-ec/ec/ecdatasource/deploymentdatasource/state"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearchstate"
)

// DataSource returns the ec_deployment data source schema.
func DataSource() *schema.Resource {
	return &schema.Resource{
		Read: read,

		Schema: newSchema(),

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func read(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.API)
	deploymentID := d.Get("id").(string)

	res, err := deploymentapi.Get(deploymentapi.GetParams{
		API:          client,
		DeploymentID: deploymentID,
		QueryParams: deputil.QueryParams{
			ShowPlans:        true,
			ShowSettings:     true,
			ShowMetadata:     true,
			ShowPlanDefaults: true,
		},
	})
	if err != nil {
		return multierror.NewPrefixed("failed retrieving deployment information", err)
	}

	d.SetId(deploymentID)

	return modelToState(d, res)
}

func modelToState(d *schema.ResourceData, res *models.DeploymentGetResponse) error {
	if err := d.Set("name", res.Name); err != nil {
		return err
	}

	if err := d.Set("healthy", res.Healthy); err != nil {
		return err
	}

	es := res.Resources.Elasticsearch[0]

	if es.Region != nil {
		if err := d.Set("region", *es.Region); err != nil {
			return err
		}
	}

	if !elasticsearchstate.IsCurrentPlanEmpty(es) {
		if err := d.Set("deployment_template_id",
			*es.Info.PlanInfo.Current.Plan.DeploymentTemplate.ID); err != nil {
			return err
		}
	}

	elasticsearchFlattened := state.FlattenElasticsearchResources(res.Resources.Elasticsearch)
	if err := d.Set("elasticsearch", elasticsearchFlattened); err != nil {
		return err
	}

	kibanaFlattened := state.FlattenKibanaResources(res.Resources.Kibana)
	if err := d.Set("kibana", kibanaFlattened); err != nil {
		return err
	}

	apmFlattened := state.FlattenApmResources(res.Resources.Apm)
	if err := d.Set("apm", apmFlattened); err != nil {
		return err
	}

	enterpriseSearchFlattened := state.FlattenEnterpriseSearchResources(res.Resources.EnterpriseSearch)
	if err := d.Set("enterprise_search", enterpriseSearchFlattened); err != nil {
		return err
	}

	return nil
}
