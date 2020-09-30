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

package trafficfilterassocresource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// newSchema returns the schema for an "ec_deployment_traffic_filter_association" resource.
func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"deployment_id": {
			Type:        schema.TypeString,
			Description: `Required deployment ID where the traffic filter will be associated`,
			Required:    true,
			ForceNew:    true,
		},
		"traffic_filter_id": {
			Type:        schema.TypeString,
			Description: "Required traffic filter ruleset ID to tie to a deployment",
			Required:    true,
			ForceNew:    true,
		},
	}
}
