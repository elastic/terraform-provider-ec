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
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandModel(d *schema.ResourceData) *models.KeystoreContents {
	var secretsIface = d.Get("secrets").([]interface{})
	var secrets = make(map[string]models.KeystoreSecret)

	for _, r := range secretsIface {
		var m = r.(map[string]interface{})

		var secretName string
		if val, ok := m["setting_name"]; ok {
			secretName = val.(string)
		}

		var secret = models.KeystoreSecret{}

		if val, ok := m["as_file"]; ok {
			secret.AsFile = ec.Bool(val.(bool))
		}

		if val, ok := m["value"]; ok {
			secret.Value = val
		}

		secrets[secretName] = secret
	}

	var request = models.KeystoreContents{
		Secrets: secrets,
	}

	return &request
}
