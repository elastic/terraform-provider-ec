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
	"time"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/plan"
	"github.com/elastic/cloud-sdk-go/pkg/plan/planutil"
)

const (
	defaultPollPlanFrequency = 2 * time.Second
	defaultMaxPlanRetry      = 4
)

// WaitForPlanCompletion waits for a pending plan to finish.
func WaitForPlanCompletion(client *api.API, id string) error {
	return planutil.Wait(plan.TrackChangeParams{
		API: client, DeploymentID: id,
		Config: plan.TrackFrequencyConfig{
			PollFrequency: defaultPollPlanFrequency,
			MaxRetries:    defaultMaxPlanRetry,
		},
	})
}
