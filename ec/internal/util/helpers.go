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
	"os"
	"strconv"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

// used in tests
var GetEnv = os.Getenv

// IsCurrentEsPlanEmpty checks that the elasticsearch resource current plan is empty.
func IsCurrentEsPlanEmpty(res *models.ElasticsearchResourceInfo) bool {
	return res.Info == nil || res.Info.PlanInfo == nil ||
		res.Info.PlanInfo.Current == nil ||
		res.Info.PlanInfo.Current.Plan == nil
}

// IsCurrentKibanaPlanEmpty checks the kibana resource current plan is empty.
func IsCurrentKibanaPlanEmpty(res *models.KibanaResourceInfo) bool {
	var emptyPlanInfo = res.Info == nil || res.Info.PlanInfo == nil || res.Info.PlanInfo.Current == nil
	return emptyPlanInfo || res.Info.PlanInfo.Current.Plan == nil
}

// IsCurrentApmPlanEmpty checks the apm resource current plan is empty.
func IsCurrentApmPlanEmpty(res *models.ApmResourceInfo) bool {
	var emptyPlanInfo = res.Info == nil || res.Info.PlanInfo == nil || res.Info.PlanInfo.Current == nil
	return emptyPlanInfo || res.Info.PlanInfo.Current.Plan == nil
}

// IsCurrentIntegrationsServerPlanEmpty checks the IntegrationsServer resource current plan is empty.
func IsCurrentIntegrationsServerPlanEmpty(res *models.IntegrationsServerResourceInfo) bool {
	var emptyPlanInfo = res.Info == nil || res.Info.PlanInfo == nil || res.Info.PlanInfo.Current == nil
	return emptyPlanInfo || res.Info.PlanInfo.Current.Plan == nil
}

// IsCurrentEssPlanEmpty checks the enterprise search resource current plan is empty.
func IsCurrentEssPlanEmpty(res *models.EnterpriseSearchResourceInfo) bool {
	var emptyPlanInfo = res.Info == nil || res.Info.PlanInfo == nil || res.Info.PlanInfo.Current == nil
	return emptyPlanInfo || res.Info.PlanInfo.Current.Plan == nil
}

// MultiGetenvOrDefault returns the value of the first environment variable in the
// given list that has a non-empty value. If none of the environment
// variables have a value, the default value is returned.
func MultiGetenvOrDefault(keys []string, defaultValue string) string {
	for _, key := range keys {
		if value := GetEnv(key); value != "" {
			return value
		}
	}
	return defaultValue
}

func StringToBool(str string) (bool, error) {
	if str == "" {
		return false, nil
	}

	v, err := strconv.ParseBool(str)
	if err != nil {
		return false, err
	}

	return v, nil
}
