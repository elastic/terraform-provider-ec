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

package util

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

// ParseDeploymentTemplate is a test helper which parse a file by path and
// returns a models.DeploymentTemplateInfoV2.
func ParseDeploymentTemplate(t *testing.T, name string) *models.DeploymentTemplateInfoV2 {
	t.Helper()
	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var res models.DeploymentTemplateInfoV2
	if err := json.NewDecoder(f).Decode(&res); err != nil {
		t.Fatal(err)
	}

	// Enriches the elasticsearch DT with the current DT.
	if len(res.DeploymentTemplate.Resources.Elasticsearch) > 0 {
		res.DeploymentTemplate.Resources.Elasticsearch[0].Plan.DeploymentTemplate = &models.DeploymentTemplateReference{
			ID: res.ID,
		}
	}

	return &res
}

// ElasticsearchResource returns the ElaticsearchPayload from a deployment
// template.
func ElasticsearchResource(res *models.DeploymentTemplateInfoV2) *models.ElasticsearchPayload {
	return res.DeploymentTemplate.Resources.Elasticsearch[0]
}

// KibanaResource returns the KibanaPayload from a deployment
// template.
func KibanaResource(res *models.DeploymentTemplateInfoV2) *models.KibanaPayload {
	return res.DeploymentTemplate.Resources.Kibana[0]
}

// ApmResource returns the ApmPayload from a deployment
// template.
func ApmResource(res *models.DeploymentTemplateInfoV2) *models.ApmPayload {
	return res.DeploymentTemplate.Resources.Apm[0]
}

// EnterpriseSearchResource returns the EnterpriseSearchPayload from a deployment
// template.
func EnterpriseSearchResource(res *models.DeploymentTemplateInfoV2) *models.EnterpriseSearchPayload {
	return res.DeploymentTemplate.Resources.EnterpriseSearch[0]
}
