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

import "github.com/elastic/cloud-sdk-go/pkg/models"

// esResource returns the ElaticsearchPayload from a deployment
// template or an empty version of the payload.
func esResource(res *models.DeploymentTemplateInfoV2) *models.ElasticsearchPayload {
	if len(res.DeploymentTemplate.Resources.Elasticsearch) == 0 {
		return &models.ElasticsearchPayload{
			Plan: &models.ElasticsearchClusterPlan{
				Elasticsearch: &models.ElasticsearchConfiguration{},
			},
			Settings: &models.ElasticsearchClusterSettings{},
		}
	}
	return res.DeploymentTemplate.Resources.Elasticsearch[0]
}

// kibanaResource returns the KibanaPayload from a deployment
// template or an empty version of the payload.
func kibanaResource(res *models.DeploymentTemplateInfoV2) *models.KibanaPayload {
	if len(res.DeploymentTemplate.Resources.Kibana) == 0 {
		return &models.KibanaPayload{
			Plan: &models.KibanaClusterPlan{
				Kibana: &models.KibanaConfiguration{},
			},
		}
	}
	return res.DeploymentTemplate.Resources.Kibana[0]
}

// apmResource returns the ApmPayload from a deployment
// template or an empty version of the payload.
func apmResource(res *models.DeploymentTemplateInfoV2) *models.ApmPayload {
	if len(res.DeploymentTemplate.Resources.Apm) == 0 {
		return &models.ApmPayload{
			Plan: &models.ApmPlan{
				Apm: &models.ApmConfiguration{},
			},
		}
	}
	return res.DeploymentTemplate.Resources.Apm[0]
}

// essResource returns the EnterpriseSearchPayload from a deployment
// template or an empty version of the payload.
func essResource(res *models.DeploymentTemplateInfoV2) *models.EnterpriseSearchPayload {
	if len(res.DeploymentTemplate.Resources.EnterpriseSearch) == 0 {
		return &models.EnterpriseSearchPayload{
			Plan: &models.EnterpriseSearchPlan{
				EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
			},
		}
	}
	return res.DeploymentTemplate.Resources.EnterpriseSearch[0]
}
