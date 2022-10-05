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

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

// flattenObservability parses a deployment's observability settings.
func flattenObservability(settings *models.DeploymentSettings) []interface{} {
	if settings == nil || settings.Observability == nil {
		return nil
	}

	var m = make(map[string]interface{})

	// We are only accepting a single deployment ID and refID for both logs and metrics.
	// If either of them is not nil the deployment ID and refID will be filled.
	if settings.Observability.Metrics != nil {
		m["deployment_id"] = settings.Observability.Metrics.Destination.DeploymentID
		m["ref_id"] = settings.Observability.Metrics.Destination.RefID
		m["metrics"] = true
	}

	if settings.Observability.Logging != nil {
		m["deployment_id"] = settings.Observability.Logging.Destination.DeploymentID
		m["ref_id"] = settings.Observability.Logging.Destination.RefID
		m["logs"] = true
	}

	if len(m) == 0 {
		return nil
	}

	return []interface{}{m}
}

func expandObservability(raw []interface{}, client *api.API) (*models.DeploymentObservabilitySettings, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var req models.DeploymentObservabilitySettings

	for _, rawObs := range raw {
		var obs = rawObs.(map[string]interface{})

		depID, ok := obs["deployment_id"].(string)
		if !ok {
			return nil, nil
		}

		refID, ok := obs["ref_id"].(string)
		if depID == "self" {
			// For self monitoring, the refID is not mandatory
			if !ok {
				refID = ""
			}
		} else if !ok || refID == "" {
			// Since ms-77, the refID is optional.
			// To not break ECE users with older versions, we still pre-calculate the refID here
			params := deploymentapi.PopulateRefIDParams{
				Kind:         util.Elasticsearch,
				API:          client,
				DeploymentID: depID,
				RefID:        ec.String(""),
			}

			if err := deploymentapi.PopulateRefID(params); err != nil {
				return nil, fmt.Errorf("observability ref_id auto discovery: %w", err)
			}

			refID = *params.RefID
		}

		if logging, ok := obs["logs"].(bool); ok && logging {
			req.Logging = &models.DeploymentLoggingSettings{
				Destination: &models.ObservabilityAbsoluteDeployment{
					DeploymentID: ec.String(depID),
					RefID:        refID,
				},
			}
		}

		if metrics, ok := obs["metrics"].(bool); ok && metrics {
			req.Metrics = &models.DeploymentMetricsSettings{
				Destination: &models.ObservabilityAbsoluteDeployment{
					DeploymentID: ec.String(depID),
					RefID:        refID,
				},
			}
		}
	}

	return &req, nil
}
