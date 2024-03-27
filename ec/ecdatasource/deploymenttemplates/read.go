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

package deploymenttemplates

import (
	"context"
	"fmt"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func (d DataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if d.client == nil {
		response.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured API client. Please report this issue to the provider developers.",
		)
		return
	}

	var data deploymentTemplatesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	res, err := deptemplateapi.List(deptemplateapi.ListParams{
		API:                        d.client,
		MetadataFilter:             "",
		Region:                     data.Region.ValueString(),
		StackVersion:               data.StackVersion.ValueString(),
		ShowHidden:                 false,
		HideInstanceConfigurations: false,
	})

	if err != nil {
		response.Diagnostics.AddError(
			"Failed retrieving deployment template list",
			fmt.Sprintf("Failed retrieving deployment template list: %s", err),
		)
		return
	}

	showDeprecated := data.ShowDeprecated.ValueBool()
	filerById := data.Id.ValueString()

	templates := mapResponseToModel(res, showDeprecated, filerById)
	data.Templates = templates

	// Finally, set the state
	diags := response.State.Set(ctx, data)
	response.Diagnostics.Append(diags...)
}

func mapResponseToModel(response []*models.DeploymentTemplateInfoV2, showDeprecated bool, idFilter string) []deploymentTemplateModel {
	templates := make([]deploymentTemplateModel, 0, len(response))
	for _, template := range response {
		if idFilter != "" && *template.ID != idFilter {
			continue
		}

		// Templates hidden in the API are considered deprecated
		hidden := isHidden(template)
		if !showDeprecated && hidden {
			continue
		}

		instanceConfigurations := template.InstanceConfigurations
		icMap := make(map[string]models.InstanceConfigurationInfo)
		for _, ic := range instanceConfigurations {
			if ic != nil {
				icMap[ic.ID] = *ic
			}
		}

		templateDefinition := template.DeploymentTemplate
		templateModel := deploymentTemplateModel{
			ID:                 *template.ID,
			Name:               *template.Name,
			Description:        template.Description,
			MinStackVersion:    template.MinVersion,
			Deprecated:         hidden,
			Elasticsearch:      mapElasticsearch(templateDefinition, icMap),
			Kibana:             mapKibana(templateDefinition, icMap),
			EnterpriseSearch:   mapEnterpriseSearch(templateDefinition, icMap),
			Apm:                mapApm(templateDefinition, icMap),
			IntegrationsServer: mapIntegrationsServer(templateDefinition, icMap),
		}
		templates = append(templates, templateModel)
	}
	return templates
}

func mapElasticsearch(templateDefinition *models.DeploymentCreateRequest, configurations map[string]models.InstanceConfigurationInfo) *elasticsearchModel {
	if templateDefinition == nil {
		return nil
	}

	resources := templateDefinition.Resources
	if resources == nil {
		return nil
	}

	payloads := resources.Elasticsearch
	if len(payloads) == 0 {
		return nil
	}

	firstEs := payloads[0]

	if firstEs.Plan == nil {
		return nil
	}

	es := elasticsearchModel{}
	for _, element := range firstEs.Plan.ClusterTopology {
		var availableSizes []string
		ic, found := configurations[element.InstanceConfigurationID]
		if found {
			if ic.DiscreteSizes != nil {
				availableSizes = make([]string, 0)
				for _, size := range ic.DiscreteSizes.Sizes {
					availableSizes = append(availableSizes, util.MemoryToState(size))
				}
			}
		}

		size := element.Size
		topology := topologyModel{
			InstanceConfigurationId:      element.InstanceConfigurationID,
			InstanceConfigurationVersion: element.InstanceConfigurationVersion,
			DefaultSize:                  util.MemoryToStateOptional(getSizeValue(size)),
			SizeResource:                 getSizeResource(size),
			AvailableSizes:               availableSizes,
			Autoscaling:                  mapAutoscaling(element),
		}

		switch element.ID {
		case "hot_content":
			es.HotTier = &topology
		case "coordinating":
			es.CoordinatingTier = &topology
		case "master":
			es.MasterTier = &topology
		case "warm":
			es.WarmTier = &topology
		case "cold":
			es.ColdTier = &topology
		case "frozen":
			es.FrozenTier = &topology
		case "ml":
			es.MlTier = &topology
		}
	}
	return &es
}

func getSizeResource(size *models.TopologySize) *string {
	if size == nil {
		return nil
	}
	return size.Resource
}

func getSizeValue(size *models.TopologySize) *int32 {
	if size == nil {
		return nil
	}
	return size.Value
}

func mapAutoscaling(element *models.ElasticsearchClusterTopologyElement) autoscalingModel {
	model := autoscalingModel{}
	if element.AutoscalingMin != nil {
		model.MinSizeResource = element.AutoscalingMin.Resource
		model.MinSize = util.MemoryToStateOptional(element.AutoscalingMin.Value)
	}
	if element.AutoscalingMax != nil {
		model.MaxSizeResource = element.AutoscalingMax.Resource
		model.MaxSize = util.MemoryToStateOptional(element.AutoscalingMax.Value)
	}
	return model
}

func mapKibana(templateDefinition *models.DeploymentCreateRequest, icMap map[string]models.InstanceConfigurationInfo) *statelessModel {
	if templateDefinition == nil {
		return nil
	}

	resources := templateDefinition.Resources
	if resources == nil {
		return nil
	}

	payloads := resources.Kibana
	if len(payloads) == 0 {
		return nil
	}
	firstKibana := payloads[0]

	if firstKibana.Plan == nil {
		return nil
	}

	topologies := firstKibana.Plan.ClusterTopology
	if len(topologies) == 0 {
		return nil
	}
	element := topologies[0]

	var availableSizes []string
	ic, found := icMap[element.InstanceConfigurationID]
	if found {
		if ic.DiscreteSizes != nil {
			availableSizes = make([]string, 0)
			for _, size := range ic.DiscreteSizes.Sizes {
				availableSizes = append(availableSizes, util.MemoryToState(size))
			}
		}
	}

	return &statelessModel{
		InstanceConfigurationId:      element.InstanceConfigurationID,
		InstanceConfigurationVersion: element.InstanceConfigurationVersion,
		DefaultSize:                  util.MemoryToStateOptional(getSizeValue(element.Size)),
		SizeResource:                 getSizeResource(element.Size),
		AvailableSizes:               availableSizes,
	}
}

func mapEnterpriseSearch(templateDefinition *models.DeploymentCreateRequest, icMap map[string]models.InstanceConfigurationInfo) *statelessModel {
	if templateDefinition == nil {
		return nil
	}

	resources := templateDefinition.Resources
	if resources == nil {
		return nil
	}

	payloads := resources.EnterpriseSearch
	if len(payloads) == 0 {
		return nil
	}
	firstEnterpriseSearch := payloads[0]

	if firstEnterpriseSearch.Plan == nil {
		return nil
	}

	topologies := firstEnterpriseSearch.Plan.ClusterTopology
	if len(topologies) == 0 {
		return nil
	}
	element := topologies[0]

	var availableSizes []string
	ic, found := icMap[element.InstanceConfigurationID]
	if found {
		if ic.DiscreteSizes != nil {
			availableSizes = make([]string, 0)
			for _, size := range ic.DiscreteSizes.Sizes {
				availableSizes = append(availableSizes, util.MemoryToState(size))
			}
		}
	}

	return &statelessModel{
		InstanceConfigurationId:      element.InstanceConfigurationID,
		InstanceConfigurationVersion: element.InstanceConfigurationVersion,
		DefaultSize:                  util.MemoryToStateOptional(getSizeValue(element.Size)),
		SizeResource:                 getSizeResource(element.Size),
		AvailableSizes:               availableSizes,
	}
}

func mapApm(templateDefinition *models.DeploymentCreateRequest, icMap map[string]models.InstanceConfigurationInfo) *statelessModel {
	if templateDefinition == nil {
		return nil
	}

	resources := templateDefinition.Resources
	if resources == nil {
		return nil
	}

	payloads := resources.Apm
	if len(payloads) == 0 {
		return nil
	}
	firstApm := payloads[0]

	if firstApm.Plan == nil {
		return nil
	}

	topologies := firstApm.Plan.ClusterTopology
	if len(topologies) == 0 {
		return nil
	}
	element := topologies[0]

	var availableSizes []string
	ic, found := icMap[element.InstanceConfigurationID]
	if found {
		if ic.DiscreteSizes != nil {
			availableSizes = make([]string, 0)
			for _, size := range ic.DiscreteSizes.Sizes {
				availableSizes = append(availableSizes, util.MemoryToState(size))
			}
		}
	}

	return &statelessModel{
		InstanceConfigurationId:      element.InstanceConfigurationID,
		InstanceConfigurationVersion: element.InstanceConfigurationVersion,
		DefaultSize:                  util.MemoryToStateOptional(getSizeValue(element.Size)),
		SizeResource:                 getSizeResource(element.Size),
		AvailableSizes:               availableSizes,
	}
}

func mapIntegrationsServer(templateDefinition *models.DeploymentCreateRequest, icMap map[string]models.InstanceConfigurationInfo) *statelessModel {
	if templateDefinition == nil {
		return nil
	}

	resources := templateDefinition.Resources
	if resources == nil {
		return nil
	}

	payloads := resources.IntegrationsServer
	if len(payloads) == 0 {
		return nil
	}
	firstIntegrationsServer := payloads[0]

	if firstIntegrationsServer.Plan == nil {
		return nil
	}

	topologies := firstIntegrationsServer.Plan.ClusterTopology
	if len(topologies) == 0 {
		return nil
	}
	element := topologies[0]

	var availableSizes []string
	ic, found := icMap[element.InstanceConfigurationID]
	if found {
		if ic.DiscreteSizes != nil {
			availableSizes = make([]string, 0)
			for _, size := range ic.DiscreteSizes.Sizes {
				availableSizes = append(availableSizes, util.MemoryToState(size))
			}
		}
	}

	return &statelessModel{
		InstanceConfigurationId:      element.InstanceConfigurationID,
		InstanceConfigurationVersion: element.InstanceConfigurationVersion,
		DefaultSize:                  util.MemoryToStateOptional(getSizeValue(element.Size)),
		SizeResource:                 getSizeResource(element.Size),
		AvailableSizes:               availableSizes,
	}
}

func isHidden(template *models.DeploymentTemplateInfoV2) bool {
	for _, metadatum := range template.Metadata {
		if *metadatum.Key == "hidden" && *metadatum.Value == "true" {
			return true
		}
	}
	return false
}
