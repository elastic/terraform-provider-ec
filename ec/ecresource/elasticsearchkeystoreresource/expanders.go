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

func expandModel(d *schema.ResourceData, delete bool) *models.KeystoreContents {
	var secrets = make(map[string]models.KeystoreSecret)

	secretName := d.Get("setting_name").(string)

	var secret = models.KeystoreSecret{}
	secret.AsFile = ec.Bool(d.Get("as_file").(bool))

	if !delete {
		secret.Value = d.Get("value")
	}

	secrets[secretName] = secret

	var request = models.KeystoreContents{
		Secrets: secrets,
	}
	return &request
}
