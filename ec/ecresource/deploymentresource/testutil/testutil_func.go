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

package testutil

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
)

// parseDeploymentTemplate is a test helper which parse a file by path and
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

func OpenDeploymentGet(t *testing.T, name string) *models.DeploymentGetResponse {
	t.Helper()
	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var res models.DeploymentGetResponse
	if err := json.NewDecoder(f).Decode(&res); err != nil {
		t.Fatal(err)
	}
	return &res
}

func EnrichWithEmptyTopologies(tpl, want *models.ElasticsearchPayload) *models.ElasticsearchPayload {
	tpl.DisplayName = want.DisplayName
	tpl.RefID = want.RefID
	tpl.Region = want.Region
	tpl.Settings = want.Settings
	tpl.Plan.AutoscalingEnabled = want.Plan.AutoscalingEnabled
	tpl.Plan.Elasticsearch = want.Plan.Elasticsearch
	tpl.Plan.Transient = want.Plan.Transient

	for i, t := range tpl.Plan.ClusterTopology {
		for _, w := range want.Plan.ClusterTopology {
			if t.ID == w.ID {
				tpl.Plan.ClusterTopology[i] = w
			}
		}
	}

	return tpl
}

func ReaderToESPayload(t *testing.T, rc io.Reader, nr bool) *models.ElasticsearchPayload {
	t.Helper()

	var tpl models.DeploymentTemplateInfoV2
	if err := json.NewDecoder(rc).Decode(&tpl); err != nil {
		t.Fatal(err)
	}

	return utils.EnrichElasticsearchTemplate(
		tpl.DeploymentTemplate.Resources.Elasticsearch[0],
		*tpl.ID,
		"",
		nr,
	)
}
