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
	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func createResourceToModel(d *schema.ResourceData, client *api.API) (*models.DeploymentCreateRequest, error) {
	var result = models.DeploymentCreateRequest{
		Name:      d.Get("name").(string),
		Resources: &models.DeploymentCreateResources{},
	}

	dtID := d.Get("deployment_template_id").(string)
	res, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     d.Get("region").(string),
		AsList:                     true,
		HideInstanceConfigurations: true,
	})
	if err != nil {
		return nil, err
	}

	esRes, err := expandEsResources(
		d.Get("elasticsearch").([]interface{}), enrichWithDeploymentTemplate(
			res.DeploymentTemplate.Resources.Elasticsearch[0], dtID,
		),
	)
	if err != nil {
		return nil, err
	}
	result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, esRes...)

	kibanaRes, err := expandKibanaResources(
		d.Get("kibana").([]interface{}),
		res.DeploymentTemplate.Resources.Kibana[0],
	)
	if err != nil {
		return nil, err
	}
	result.Resources.Kibana = append(result.Resources.Kibana, kibanaRes...)

	apmRes, err := expandApmResources(
		d.Get("apm").([]interface{}),
		res.DeploymentTemplate.Resources.Apm[0],
	)
	if err != nil {
		return nil, err
	}
	result.Resources.Apm = append(result.Resources.Apm, apmRes...)

	enterpriseSearchRes, err := expandEssResources(
		d.Get("enterprise_search").([]interface{}),
		res.DeploymentTemplate.Resources.EnterpriseSearch[0],
	)
	if err != nil {
		return nil, err
	}
	result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchRes...)

	expandTrafficFilterCreate(d.Get("traffic_filter").(*schema.Set), &result)

	return &result, nil
}

func updateResourceToModel(d *schema.ResourceData, client *api.API) (*models.DeploymentUpdateRequest, error) {
	var result = models.DeploymentUpdateRequest{
		Name:         d.Get("name").(string),
		PruneOrphans: ec.Bool(true),
		Resources:    &models.DeploymentUpdateResources{},
	}

	dtID := d.Get("deployment_template_id").(string)
	res, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     d.Get("region").(string),
		AsList:                     true,
		HideInstanceConfigurations: true,
	})
	if err != nil {
		return nil, err
	}

	es := d.Get("elasticsearch").([]interface{})
	kibana := d.Get("kibana").([]interface{})
	apm := d.Get("apm").([]interface{})
	enterpriseSearch := d.Get("enterprise_search").([]interface{})

	// When the deployment template is changed, we need to unset all the
	// resources topologies to account for a new instance_configuration_id.
	dtChange := d.HasChange("deployment_template_id")
	if o, _ := d.GetChange("deployment_template_id"); dtChange && o.(string) != "" {
		unsetTopology(es, d)
		unsetTopology(kibana, d)
		unsetTopology(apm, d)
		unsetTopology(enterpriseSearch, d)
	}

	esRes, err := expandEsResources(
		es, enrichWithDeploymentTemplate(
			res.DeploymentTemplate.Resources.Elasticsearch[0], dtID,
		),
	)
	if err != nil {
		return nil, err
	}
	result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, esRes...)

	kibanaRes, err := expandKibanaResources(
		kibana, res.DeploymentTemplate.Resources.Kibana[0],
	)
	if err != nil {
		return nil, err
	}
	result.Resources.Kibana = append(result.Resources.Kibana, kibanaRes...)

	apmRes, err := expandApmResources(
		apm, res.DeploymentTemplate.Resources.Apm[0],
	)
	if err != nil {
		return nil, err
	}
	result.Resources.Apm = append(result.Resources.Apm, apmRes...)

	enterpriseSearchRes, err := expandEssResources(
		enterpriseSearch,
		res.DeploymentTemplate.Resources.EnterpriseSearch[0],
	)
	if err != nil {
		return nil, err
	}
	result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchRes...)

	return &result, nil
}

func enrichWithDeploymentTemplate(tpl *models.ElasticsearchPayload, dt string) *models.ElasticsearchPayload {
	if tpl.Plan.DeploymentTemplate == nil {
		tpl.Plan.DeploymentTemplate = &models.DeploymentTemplateReference{}
	}

	if tpl.Plan.DeploymentTemplate.ID == nil || *tpl.Plan.DeploymentTemplate.ID == "" {
		tpl.Plan.DeploymentTemplate.ID = ec.String(dt)
	}

	return tpl
}

func unsetTopology(rawRes []interface{}, d *schema.ResourceData) {
	for _, r := range rawRes {
		delete(r.(map[string]interface{}), "topology")
	}
}
