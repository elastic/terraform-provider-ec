package v2

import "github.com/elastic/cloud-sdk-go/pkg/models"

func removeAutoscalingTierOverridesFromTemplate(template *models.DeploymentTemplateInfoV2) {
	for _, esResource := range template.DeploymentTemplate.Resources.Elasticsearch {
		for _, topologyElem := range esResource.Plan.ClusterTopology {
			topologyElem.AutoscalingTierOverride = nil
		}
	}
}
