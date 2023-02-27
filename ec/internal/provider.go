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

package internal

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/elastic/cloud-sdk-go/pkg/api"
)

// ConvertProviderData is a helper function for DataSource.Configure and Resource.Configure implementations
func ConvertProviderData(providerData any) (*api.API, diag.Diagnostics) {
	var diags diag.Diagnostics

	if providerData == nil {
		return nil, diags
	}

	client, ok := providerData.(*api.API)
	if !ok {
		diags.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *api.API, got: %T. Please report this issue to the provider developers.", providerData),
		)

		return nil, diags
	}
	return client, diags
}
