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
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

// parseDeploymentTemplate is a test helper which parse a file by path and
// returns a models.DeploymentTemplateInfoV2.
func parseDeploymentTemplate(t *testing.T, name string) *models.DeploymentTemplateInfoV2 {
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

// esResource returns the ElaticsearchPayload from a deployment
// template.
func esResource(res *models.DeploymentTemplateInfoV2) *models.ElasticsearchPayload {
	return res.DeploymentTemplate.Resources.Elasticsearch[0]
}

// kibanaResource returns the KibanaPayload from a deployment
// template.
func kibanaResource(res *models.DeploymentTemplateInfoV2) *models.KibanaPayload {
	return res.DeploymentTemplate.Resources.Kibana[0]
}

// apmResource returns the ApmPayload from a deployment
// template.
func apmResource(res *models.DeploymentTemplateInfoV2) *models.ApmPayload {
	return res.DeploymentTemplate.Resources.Apm[0]
}

// essResource returns the EnterpriseSearchPayload from a deployment
// template.
func essResource(res *models.DeploymentTemplateInfoV2) *models.EnterpriseSearchPayload {
	return res.DeploymentTemplate.Resources.EnterpriseSearch[0]
}

// flattenClusterEndpoint receives a ClusterMetadataInfo, parses the http and
// https endpoints and returns a map with two keys: `http_endpoint` and
// `https_endpoint`
func flattenClusterEndpoint(metadata *models.ClusterMetadataInfo) map[string]interface{} {
	if metadata == nil || metadata.Endpoint == "" || metadata.Ports == nil {
		return nil
	}

	var m = make(map[string]interface{})
	if metadata.Ports.HTTP != nil {
		m["http_endpoint"] = fmt.Sprintf("http://%s:%d", metadata.Endpoint, *metadata.Ports.HTTP)
	}

	if metadata.Ports.HTTPS != nil {
		m["https_endpoint"] = fmt.Sprintf("https://%s:%d", metadata.Endpoint, *metadata.Ports.HTTPS)
	}

	return m
}
