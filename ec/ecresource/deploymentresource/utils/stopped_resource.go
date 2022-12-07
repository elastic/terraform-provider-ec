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

import "github.com/elastic/cloud-sdk-go/pkg/models"

// IsApmResourceStopped returns true if the resource is stopped.
func IsApmResourceStopped(res *models.ApmResourceInfo) bool {
	return res == nil || res.Info == nil || res.Info.Status == nil ||
		*res.Info.Status == "stopped"
}

// IsIntegrationsServerResourceStopped returns true if the resource is stopped.
func IsIntegrationsServerResourceStopped(res *models.IntegrationsServerResourceInfo) bool {
	return res == nil || res.Info == nil || res.Info.Status == nil ||
		*res.Info.Status == "stopped"
}

// IsEsResourceStopped returns true if the resource is stopped.
func IsEsResourceStopped(res *models.ElasticsearchResourceInfo) bool {
	return res == nil || res.Info == nil || res.Info.Status == nil ||
		*res.Info.Status == "stopped"
}

// IsEssResourceStopped returns true if the resource is stopped.
func IsEssResourceStopped(res *models.EnterpriseSearchResourceInfo) bool {
	return res == nil || res.Info == nil || res.Info.Status == nil ||
		*res.Info.Status == "stopped"
}

// IsKibanaResourceStopped returns true if the resource is stopped.
func IsKibanaResourceStopped(res *models.KibanaResourceInfo) bool {
	return res == nil || res.Info == nil || res.Info.Status == nil ||
		*res.Info.Status == "stopped"
}
