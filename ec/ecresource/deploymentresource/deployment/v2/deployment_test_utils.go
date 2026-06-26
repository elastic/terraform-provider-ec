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

package v2

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/models"

	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
)

func elasticsearchPayloadFromReader(t *testing.T, rc io.Reader, useNodeRoles bool) *models.ElasticsearchPayload {
	t.Helper()

	var tpl models.DeploymentTemplateInfoV2
	if err := json.NewDecoder(rc).Decode(&tpl); err != nil {
		t.Fatal(err)
	}

	return elasticsearchv2.EnrichElasticsearchTemplate(
		tpl.DeploymentTemplate.Resources.Elasticsearch[0],
		*tpl.ID,
		"",
		useNodeRoles,
	)
}

func deploymentGetResponseFromFile(t *testing.T, filename string) *models.DeploymentGetResponse {
	t.Helper()
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	var res models.DeploymentGetResponse
	if err := json.NewDecoder(f).Decode(&res); err != nil {
		t.Fatal(err)
	}
	return &res
}
