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

package deploymentresource

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/apmstate"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearchstate"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/kibanastate"
)

// TODO: EnterpriseSearch
// TODO: AppSearch ? EnterpriseSearch superseeds AppSearch, might not be worth spending time
// on it.
func createResourceToModel(d *schema.ResourceData) (*models.DeploymentCreateRequest, error) {
	var result = models.DeploymentCreateRequest{
		Name: d.Get("name").(string),
		Resources: &models.DeploymentCreateResources{
			Apm:              make([]*models.ApmPayload, 0),
			Appsearch:        make([]*models.AppSearchPayload, 0),
			Elasticsearch:    make([]*models.ElasticsearchPayload, 0),
			EnterpriseSearch: make([]*models.EnterpriseSearchPayload, 0),
			Kibana:           make([]*models.KibanaPayload, 0),
		},
	}

	esRes, err := elasticsearchstate.ExpandResources(
		d.Get("elasticsearch").([]interface{}),
		d.Get("deployment_template_id").(string),
	)
	if err != nil {
		return nil, err
	}
	result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, esRes...)

	kibanaRes, err := kibanastate.ExpandResources(d.Get("kibana").([]interface{}))
	if err != nil {
		return nil, err
	}
	result.Resources.Kibana = append(result.Resources.Kibana, kibanaRes...)

	apmRes, err := apmstate.ExpandResources(d.Get("apm").([]interface{}))
	if err != nil {
		return nil, err
	}
	result.Resources.Apm = append(result.Resources.Apm, apmRes...)

	return &result, nil
}

// TODO: EnterpriseSearch
// TODO: AppSearch ? EnterpriseSearch superseeds AppSearch, might not be worth spending time
// on it.
func updateResourceToModel(d *schema.ResourceData) (*models.DeploymentUpdateRequest, error) {
	var result = models.DeploymentUpdateRequest{
		Name: d.Get("name").(string),
		// Setting this to false since we might not support all API resources in
		// the provivider, setting to true, might cause some resources to be set
		// incorrectly to "[]", which will cause the resources to be deleted.
		PruneOrphans: ec.Bool(false),
		Resources: &models.DeploymentUpdateResources{
			Apm:              make([]*models.ApmPayload, 0),
			Appsearch:        make([]*models.AppSearchPayload, 0),
			Elasticsearch:    make([]*models.ElasticsearchPayload, 0),
			EnterpriseSearch: make([]*models.EnterpriseSearchPayload, 0),
			Kibana:           make([]*models.KibanaPayload, 0),
		},
	}

	esRes, err := elasticsearchstate.ExpandResources(
		d.Get("elasticsearch").([]interface{}),
		d.Get("deployment_template_id").(string),
	)
	if err != nil {
		return nil, err
	}
	result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, esRes...)

	kibanaRes, err := kibanastate.ExpandResources(d.Get("kibana").([]interface{}))
	if err != nil {
		return nil, err
	}
	result.Resources.Kibana = append(result.Resources.Kibana, kibanaRes...)

	apmRes, err := apmstate.ExpandResources(d.Get("apm").([]interface{}))
	if err != nil {
		return nil, err
	}
	result.Resources.Apm = append(result.Resources.Apm, apmRes...)

	return &result, nil
}
