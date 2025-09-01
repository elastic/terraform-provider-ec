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
	"context"
	"time"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"
	"github.com/elastic/cloud-sdk-go/pkg/plan"
	"github.com/elastic/cloud-sdk-go/pkg/plan/planutil"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
)

const (
	defaultPollPlanFrequency      = 2 * time.Second
	defaultMaxPlanRetry           = 4
	defaultIntegrationsServerWait = 2 * time.Minute
)

// WaitForPlanCompletion waits for a pending plan to finish.
func WaitForPlanCompletion(client *api.API, id string) error {
	err := planutil.Wait(plan.TrackChangeParams{
		API: client, DeploymentID: id,
		Config: plan.TrackFrequencyConfig{
			PollFrequency: defaultPollPlanFrequency,
			MaxRetries:    defaultMaxPlanRetry,
		},
	})
	if err != nil {
		return err
	}

	return waitForIntegrationServerEndpoints(client, id)
}

func waitForIntegrationServerEndpoints(client *api.API, id string) error {
	timeout, cancel := context.WithTimeout(context.Background(), defaultIntegrationsServerWait)
	defer cancel()
	for {
		err := timeout.Err()
		if err != nil {
			return err
		}

		response, err := deploymentapi.Get(deploymentapi.GetParams{
			API:          client,
			DeploymentID: id,
			QueryParams: deputil.QueryParams{
				ShowSettings:               true,
				ShowPlans:                  true,
				ShowMetadata:               true,
				ShowPlanDefaults:           true,
				ShowInstanceConfigurations: true,
			},
		})
		if err != nil {
			return err
		}

		if len(response.Resources.IntegrationsServer) == 0 {
			return nil
		}

		for _, intSrvr := range response.Resources.IntegrationsServer {
			if integrationsserverv2.IsIntegrationsServerStopped(intSrvr) {
				return nil
			}

			if intSrvr.Info == nil {
				continue
			}

			if intSrvr.Info.Metadata == nil {
				continue
			}

			isStarted := intSrvr.Info.Status != nil && *intSrvr.Info.Status == "started"
			hasServiceUrls := len(intSrvr.Info.Metadata.ServicesUrls) > 0

			if isStarted && hasServiceUrls {
				return nil
			}
		}
	}
}
