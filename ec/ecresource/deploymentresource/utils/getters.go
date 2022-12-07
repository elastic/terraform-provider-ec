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

package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func HasRunningResources(res *models.DeploymentGetResponse) bool {
	var hasRunning bool
	if res.Resources != nil {
		for _, r := range res.Resources.Elasticsearch {
			if !IsEsResourceStopped(r) {
				hasRunning = true
			}
		}
		for _, r := range res.Resources.Kibana {
			if !IsKibanaResourceStopped(r) {
				hasRunning = true
			}
		}
		for _, r := range res.Resources.Apm {
			if !IsApmResourceStopped(r) {
				hasRunning = true
			}
		}
		for _, r := range res.Resources.EnterpriseSearch {
			if !IsEssResourceStopped(r) {
				hasRunning = true
			}
		}
		for _, r := range res.Resources.IntegrationsServer {
			if !IsIntegrationsServerResourceStopped(r) {
				hasRunning = true
			}
		}
	}
	return hasRunning
}

func GetDeploymentTemplateID(res *models.DeploymentResources) (string, error) {
	var deploymentTemplateID string
	var foundTemplates []string
	for _, esRes := range res.Elasticsearch {
		if util.IsCurrentEsPlanEmpty(esRes) {
			continue
		}

		var emptyDT = esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate == nil
		if emptyDT {
			continue
		}

		if deploymentTemplateID == "" {
			deploymentTemplateID = *esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate.ID
		}

		foundTemplates = append(foundTemplates,
			*esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate.ID,
		)
	}

	if deploymentTemplateID == "" {
		return "", errors.New("failed to obtain the deployment template id")
	}

	if len(foundTemplates) > 1 {
		return "", fmt.Errorf(
			"there are more than 1 deployment templates specified on the deployment: \"%s\"", strings.Join(foundTemplates, ", "),
		)
	}

	return deploymentTemplateID, nil
}

func GetRegion(res *models.DeploymentResources) (region string) {
	for _, r := range res.Elasticsearch {
		if r.Region != nil && *r.Region != "" {
			return *r.Region
		}
	}

	return region
}

func GetLowestVersion(res *models.DeploymentResources) (string, error) {
	// We're starting off with a very high version so it can be replaced.
	replaceVersion := `99.99.99`
	version := semver.MustParse(replaceVersion)
	for _, r := range res.Elasticsearch {
		if !util.IsCurrentEsPlanEmpty(r) {
			v := r.Info.PlanInfo.Current.Plan.Elasticsearch.Version
			if err := swapLowerVersion(&version, v); err != nil && !IsEsResourceStopped(r) {
				return "", fmt.Errorf("elasticsearch version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.Kibana {
		if !util.IsCurrentKibanaPlanEmpty(r) {
			v := r.Info.PlanInfo.Current.Plan.Kibana.Version
			if err := swapLowerVersion(&version, v); err != nil && !IsKibanaResourceStopped(r) {
				return version.String(), fmt.Errorf("kibana version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.Apm {
		if !util.IsCurrentApmPlanEmpty(r) {
			v := r.Info.PlanInfo.Current.Plan.Apm.Version
			if err := swapLowerVersion(&version, v); err != nil && !IsApmResourceStopped(r) {
				return version.String(), fmt.Errorf("apm version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.IntegrationsServer {
		if !util.IsCurrentIntegrationsServerPlanEmpty(r) {
			v := r.Info.PlanInfo.Current.Plan.IntegrationsServer.Version
			if err := swapLowerVersion(&version, v); err != nil && !IsIntegrationsServerResourceStopped(r) {
				return version.String(), fmt.Errorf("integrations_server version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.EnterpriseSearch {
		if !util.IsCurrentEssPlanEmpty(r) {
			v := r.Info.PlanInfo.Current.Plan.EnterpriseSearch.Version
			if err := swapLowerVersion(&version, v); err != nil && !IsEssResourceStopped(r) {
				return version.String(), fmt.Errorf("enterprise search version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	if version.String() != replaceVersion {
		return version.String(), nil
	}
	return "", errors.New("unable to determine the lowest version for any the deployment components")
}

func swapLowerVersion(version *semver.Version, comp string) error {
	if comp == "" {
		return nil
	}

	v, err := semver.Parse(comp)
	if err != nil {
		return err
	}
	if v.LT(*version) {
		*version = v
	}
	return nil
}
