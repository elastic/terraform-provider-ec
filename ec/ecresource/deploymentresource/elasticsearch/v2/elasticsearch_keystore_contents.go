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

package v2

import (
	"context"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchKeystoreContentsTF struct {
	Value  types.String `tfsdk:"value"`
	AsFile types.Bool   `tfsdk:"as_file"`
}

type ElasticsearchKeystoreContents struct {
	Value  string `tfsdk:"value"`
	AsFile *bool  `tfsdk:"as_file"`
}

func elasticsearchKeystoreContentsPayload(ctx context.Context, keystoreContentsTF types.Map, model *models.ElasticsearchClusterSettings, esState *ElasticsearchTF) (*models.ElasticsearchClusterSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	if (keystoreContentsTF.IsNull() || len(keystoreContentsTF.Elements()) == 0) && esState == nil {
		return model, nil
	}

	secrets := make(map[string]models.KeystoreSecret, len(keystoreContentsTF.Elements()))

	for secretKey, elem := range keystoreContentsTF.Elements() {
		var secretTF ElasticsearchKeystoreContentsTF

		ds := tfsdk.ValueAs(ctx, elem, &secretTF)
		diags.Append(ds...)

		if ds.HasError() {
			continue
		}

		var secret models.KeystoreSecret

		secret.AsFile = ec.Bool(false)

		if !secretTF.AsFile.IsUnknown() && !secretTF.AsFile.IsNull() {
			secret.AsFile = ec.Bool(secretTF.AsFile.ValueBool())
		}
		secret.Value = secretTF.Value.ValueString()

		secrets[secretKey] = secret
	}

	// remove secrets that were in state but are removed from plan
	if esState != nil {
		if !esState.KeystoreContents.IsNull() {
			for k := range esState.KeystoreContents.Elements() {
				if _, ok := secrets[k]; !ok {
					secrets[k] = models.KeystoreSecret{}
				}
			}
		}
	}

	if len(secrets) == 0 {
		return model, nil
	}

	if model == nil {
		model = &models.ElasticsearchClusterSettings{}
	}

	if model.KeystoreContents == nil {
		model.KeystoreContents = new(models.KeystoreContents)
	}

	model.KeystoreContents.Secrets = secrets

	return model, nil
}
