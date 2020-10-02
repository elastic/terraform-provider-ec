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

package elasticsearchkeystoreresource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"deployment_id": {
			Type:        schema.TypeString,
			Description: `Required deployment ID corresponding to the Elasticsearch resource keystore`,
			Required:    true,
		},
		"secrets": {
			Type:        schema.TypeList,
			Description: "",
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"setting_name": {
						Type:        schema.TypeString,
						Description: "Required name for the setting. Must be unique",
						Required:    true,
					},
					"value": {
						Type:        schema.TypeString,
						Description: "Value of this setting. This can either be a string or a JSON object that is stored as a JSON string in the keystore.",
						Optional:    true,
					},
					"as_file": {
						Type:        schema.TypeBool,
						Description: "Stores the keystore secret as a file. The default is false, which stores the keystore secret as string when value is a plain string",
						Computed:    true,
						Optional:    true,
					},
				},
			},
		},
	}
}
