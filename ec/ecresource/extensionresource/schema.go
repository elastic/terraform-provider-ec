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

package extensionresource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "Required name of the ruleset",
			Required:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Description for extension",
			Optional:    true,
		},
		"extension_type": {
			Type:        schema.TypeString,
			Description: "Extension type. bundle or plugin",
			Required:    true,
		},
		"version": {
			Type:        schema.TypeString,
			Description: "Eleasticsearch version",
			Required:    true,
		},
		"download_url": {
			Type:        schema.TypeString,
			Description: "download url",
			Optional:    true,
		},

		// Uploading file via API
		"file_path": {
			Type:         schema.TypeString,
			Description:  "file path",
			Optional:     true,
			RequiredWith: []string{"file_hash"},
		},
		"file_hash": {
			Type:        schema.TypeString,
			Description: "file hash",
			Optional:    true,
		},

		"url": {
			Type:        schema.TypeString,
			Description: "",
			Computed:    true,
		},
		"last_modified": {
			Type:        schema.TypeString,
			Description: "",
			Computed:    true,
		},
		"size": {
			Type:        schema.TypeInt,
			Description: "",
			Computed:    true,
		},
	}
}
