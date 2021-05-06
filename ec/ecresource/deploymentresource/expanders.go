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
	"fmt"
	"sort"

	"github.com/blang/semver/v4"
	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func createResourceToModel(d *schema.ResourceData, client *api.API) (*models.DeploymentCreateRequest, error) {
	var result = models.DeploymentCreateRequest{
		Name:      d.Get("name").(string),
		Alias:     d.Get("alias").(string),
		Resources: &models.DeploymentCreateResources{},
		Settings:  &models.DeploymentCreateSettings{},
		Metadata:  &models.DeploymentCreateMetadata{},
	}

	dtID := d.Get("deployment_template_id").(string)
	version := d.Get("version").(string)
	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     d.Get("region").(string),
		HideInstanceConfigurations: true,
	})
	if err != nil {
		return nil, err
	}

	useNodeRoles, err := compatibleWithNodeRoles(version)
	if err != nil {
		return nil, err
	}

	merr := multierror.NewPrefixed("invalid configuration")
	esRes, err := expandEsResources(
		d.Get("elasticsearch").([]interface{}),
		enrichElasticsearchTemplate(
			esResource(template), dtID, version, useNodeRoles,
		),
	)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, esRes...)

	kibanaRes, err := expandKibanaResources(
		d.Get("kibana").([]interface{}), kibanaResource(template),
	)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.Kibana = append(result.Resources.Kibana, kibanaRes...)

	apmRes, err := expandApmResources(
		d.Get("apm").([]interface{}), apmResource(template),
	)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.Apm = append(result.Resources.Apm, apmRes...)

	enterpriseSearchRes, err := expandEssResources(
		d.Get("enterprise_search").([]interface{}), essResource(template),
	)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchRes...)

	if err := merr.ErrorOrNil(); err != nil {
		return nil, err
	}

	expandTrafficFilterCreate(d.Get("traffic_filter").(*schema.Set), &result)

	observability, err := expandObservability(d.Get("observability").([]interface{}), client)
	if err != nil {
		return nil, err
	}
	result.Settings.Observability = observability

	result.Metadata.Tags = expandTags(d.Get("tags").(map[string]interface{}))

	return &result, nil
}

func updateResourceToModel(d *schema.ResourceData, client *api.API) (*models.DeploymentUpdateRequest, error) {
	var result = models.DeploymentUpdateRequest{
		Name:         d.Get("name").(string),
		PruneOrphans: ec.Bool(true),
		Resources:    &models.DeploymentUpdateResources{},
		Settings:     &models.DeploymentUpdateSettings{},
		Metadata:     &models.DeploymentUpdateMetadata{},
	}

	dtID := d.Get("deployment_template_id").(string)
	version := d.Get("version").(string)
	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     d.Get("region").(string),
		HideInstanceConfigurations: true,
	})
	if err != nil {
		return nil, err
	}

	es := d.Get("elasticsearch").([]interface{})
	kibana := d.Get("kibana").([]interface{})
	apm := d.Get("apm").([]interface{})
	enterpriseSearch := d.Get("enterprise_search").([]interface{})

	// When the deployment template is changed, we need to unset the missing
	// resource topologies to account for a new instance_configuration_id and
	// a different default value.
	prevDT, _ := d.GetChange("deployment_template_id")
	if d.HasChange("deployment_template_id") && prevDT.(string) != "" {
		// If the deployment_template_id is changed, then we unset the
		// Elasticsearch topology to account for the case where the
		// instance_configuration_id changes, i.e. Hot / Warm, etc.

		// This might not be necessary going forward as we move to
		// tiered Elasticsearch nodes.
		unsetTopology(es)
	}

	nodeRolesCompatible, err := compatibleWithNodeRoles(version)
	if err != nil {
		return nil, err
	}
	useNodeRoles := d.HasChange("elasticsearch") && !d.HasChange("version") && nodeRolesCompatible

	merr := multierror.NewPrefixed("invalid configuration")
	esRes, err := expandEsResources(
		es, enrichElasticsearchTemplate(
			esResource(template), dtID, version, useNodeRoles,
		),
	)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, esRes...)

	kibanaRes, err := expandKibanaResources(kibana, kibanaResource(template))
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.Kibana = append(result.Resources.Kibana, kibanaRes...)

	apmRes, err := expandApmResources(apm, apmResource(template))
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.Apm = append(result.Resources.Apm, apmRes...)

	enterpriseSearchRes, err := expandEssResources(enterpriseSearch, essResource(template))
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchRes...)

	if err := merr.ErrorOrNil(); err != nil {
		return nil, err
	}

	observability, err := expandObservability(d.Get("observability").([]interface{}), client)
	if err != nil {
		return nil, err
	}
	result.Settings.Observability = observability

	// In order to stop shipping logs and metrics, an empty Observability
	// object must be passed, as opposed to a nil object when creating a
	// deployment without observability settings.
	if util.ObjectRemoved(d, "observability") {
		result.Settings.Observability = &models.DeploymentObservabilitySettings{}
	}

	result.Metadata.Tags = expandTags(d.Get("tags").(map[string]interface{}))

	return &result, nil
}

func enrichElasticsearchTemplate(tpl *models.ElasticsearchPayload, dt, version string, useNodeRoles bool) *models.ElasticsearchPayload {
	if tpl.Plan.DeploymentTemplate == nil {
		tpl.Plan.DeploymentTemplate = &models.DeploymentTemplateReference{}
	}

	if tpl.Plan.DeploymentTemplate.ID == nil || *tpl.Plan.DeploymentTemplate.ID == "" {
		tpl.Plan.DeploymentTemplate.ID = ec.String(dt)
	}

	if tpl.Plan.Elasticsearch.Version == "" {
		tpl.Plan.Elasticsearch.Version = version
	}

	for _, topology := range tpl.Plan.ClusterTopology {
		if useNodeRoles {
			topology.NodeType = nil
			continue
		}
		topology.NodeRoles = nil
	}

	return tpl
}

func unsetTopology(rawRes []interface{}) {
	for _, r := range rawRes {
		delete(r.(map[string]interface{}), "topology")
	}
}

func expandTags(raw map[string]interface{}) []*models.MetadataItem {
	result := make([]*models.MetadataItem, 0, len(raw))
	for k, v := range raw {
		result = append(result, &models.MetadataItem{
			Key:   ec.String(k),
			Value: ec.String(v.(string)),
		})
	}

	// Sort by key
	sort.SliceStable(result, func(i, j int) bool {
		return *result[i].Key < *result[j].Key
	})

	return result
}

func compatibleWithNodeRoles(version string) (bool, error) {
	deploymentVersion, err := semver.Parse(version)
	if err != nil {
		return false, fmt.Errorf("failed to parse Elasticsearch version: %w", err)
	}

	dataTiersVersion := semver.MustParse("7.10.0")
	return deploymentVersion.GE(dataTiersVersion), nil
}
