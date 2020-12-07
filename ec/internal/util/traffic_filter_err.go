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
	"errors"
	"net/http"

	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments_traffic_filter"
)

// TrafficFilterNotFound returns true when the error is a 404 or 403.
func TrafficFilterNotFound(err error) bool {
	// We're using the As() call since we do not care about the error value
	// but do care about the error type since it's an implicit 404.
	var notDeploymentNotFound *deployments_traffic_filter.GetTrafficFilterRulesetNotFound
	if errors.As(err, &notDeploymentNotFound) {
		return true
	}

	// We also check for the case where a 403 is thrown for ESS.
	return apierror.IsRuntimeStatusCode(err, http.StatusForbidden)
}
