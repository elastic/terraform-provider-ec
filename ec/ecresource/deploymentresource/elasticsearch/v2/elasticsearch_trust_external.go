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
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTrustExternals v1.ElasticsearchTrustExternals

func ReadElasticsearchTrustExternals(in *models.ElasticsearchClusterSettings) (ElasticsearchTrustExternals, error) {
	if in == nil || in.Trust == nil {
		return nil, nil
	}

	externals := make(ElasticsearchTrustExternals, 0, len(in.Trust.External))

	for _, model := range in.Trust.External {
		external, err := ReadElasticsearchTrustExternal(model)
		if err != nil {
			return nil, err
		}
		externals = append(externals, *external)
	}

	return externals, nil
}

func ElasticsearchTrustExternalPayload(ctx context.Context, externals types.Set, model *models.ElasticsearchClusterSettings) (*models.ElasticsearchClusterSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	payloads := make([]*models.ExternalTrustRelationship, 0, len(externals.Elems))

	for _, elem := range externals.Elems {
		var external v1.ElasticsearchTrustExternalTF

		ds := tfsdk.ValueAs(ctx, elem, &external)

		diags = append(diags, ds...)

		if diags.HasError() {
			continue
		}

		id := external.RelationshipId.Value
		all := external.TrustAll.Value

		payload := &models.ExternalTrustRelationship{
			TrustRelationshipID: &id,
			TrustAll:            &all,
		}

		ds = external.TrustAllowlist.ElementsAs(ctx, &payload.TrustAllowlist, true)

		diags = append(diags, ds...)

		if ds.HasError() {
			continue
		}

		payloads = append(payloads, payload)
	}

	if len(payloads) == 0 {
		return model, nil
	}

	if model == nil {
		model = &models.ElasticsearchClusterSettings{}
	}

	if model.Trust == nil {
		model.Trust = &models.ElasticsearchClusterTrustSettings{}
	}

	model.Trust.External = append(model.Trust.External, payloads...)

	return model, nil
}

func ReadElasticsearchTrustExternal(in *models.ExternalTrustRelationship) (*v1.ElasticsearchTrustExternal, error) {
	var ext v1.ElasticsearchTrustExternal

	if in.TrustRelationshipID != nil {
		ext.RelationshipId = in.TrustRelationshipID
	}

	if in.TrustAll != nil {
		ext.TrustAll = in.TrustAll
	}

	ext.TrustAllowlist = in.TrustAllowlist

	return &ext, nil
}

func ElasticsearchTrustAccountPayload(ctx context.Context, accounts types.Set, model *models.ElasticsearchClusterSettings) (*models.ElasticsearchClusterSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	payloads := make([]*models.AccountTrustRelationship, 0, len(accounts.Elems))

	for _, elem := range accounts.Elems {
		var account v1.ElasticsearchTrustAccountTF

		ds := tfsdk.ValueAs(ctx, elem, &account)

		diags = append(diags, ds...)

		if ds.HasError() {
			continue
		}

		id := account.AccountId.Value
		all := account.TrustAll.Value

		payload := &models.AccountTrustRelationship{
			AccountID: &id,
			TrustAll:  &all,
		}

		ds = account.TrustAllowlist.ElementsAs(ctx, &payload.TrustAllowlist, true)

		diags = append(diags, ds...)

		if ds.HasError() {
			continue
		}

		payloads = append(payloads, payload)
	}

	if len(payloads) == 0 {
		return model, nil
	}

	if model == nil {
		model = &models.ElasticsearchClusterSettings{}
	}

	if model.Trust == nil {
		model.Trust = &models.ElasticsearchClusterTrustSettings{}
	}

	model.Trust.Accounts = append(model.Trust.Accounts, payloads...)

	return model, nil
}
