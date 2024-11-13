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

package acc

import (
	"fmt"
	"os"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/stackapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/slice"
)

const (
	defaultTemplate        = "storage-optimized"
	generalPurposeTemplate = "general-purpose"
	cpuOpTemplate          = "cpu-optimized"
	vectorSearchTemplate   = "vector-search-optimized"
)

func getRegion() string {
	region := "us-east-1"

	if r := os.Getenv("EC_REGION"); r != "" {
		region = r
	}

	return region
}

func latestStackVersion() (string, error) {
	client, err := newAPI()
	if err != nil {
		return "", err
	}

	res, err := stackapi.List(stackapi.ListParams{
		API:    client,
		Region: getRegion(),
	})
	if err != nil {
		return "", err
	}

	return res.Stacks[0].Version, nil
}

func setDefaultTemplate(region, template string) string {
	if strings.Contains(region, "azure") {
		region = "azure"
	}

	if strings.Contains(region, "gcp") {
		region = "gcp"
	}

	switch region {
	case "azure":
		return "azure-" + template
	case "gcp":
		return "gcp-" + template
	default:
		return buildAwsTemplate(template)
	}
}

func buildAwsTemplate(template string) string {
	armTemplates := []string{
		vectorSearchTemplate,
		cpuOpTemplate,
	}

	if slice.HasString(armTemplates, template) {
		return "aws-" + template + "-arm"
	}

	return "aws-" + template
}

func getResources(deploymentTemplate string) (*models.DeploymentCreateResources, error) {
	client, err := newAPI()
	if err != nil {
		return nil, err
	}

	res, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:        client,
		TemplateID: deploymentTemplate,
		Region:     getRegion(),
	})
	if err != nil {
		return nil, err
	}

	return res.DeploymentTemplate.Resources, nil
}

func setInstanceConfigurations(deploymentTemplate string) (esIC, kibanaIC, apmIC, essIC string, err error) {
	resources, err := getResources(deploymentTemplate)
	if err != nil {
		return "", "", "", "", err
	}

	esRes := resources.Elasticsearch[0].Plan.ClusterTopology

	for _, t := range esRes {
		if *t.Size.Value > 0 {
			esIC = t.InstanceConfigurationID
		}
	}

	if esIC == "" {
		return "", "", "", "",
			fmt.Errorf(
				"could not find default instance configuration for Elasticsearch, verify  details for: %v",
				deploymentTemplate)
	}

	kibanaIC = resources.Kibana[0].
		Plan.ClusterTopology[0].InstanceConfigurationID

	apmIC = resources.Apm[0].
		Plan.ClusterTopology[0].InstanceConfigurationID

	essIC = resources.EnterpriseSearch[0].
		Plan.ClusterTopology[0].InstanceConfigurationID

	return esIC, kibanaIC, apmIC, essIC, nil
}
