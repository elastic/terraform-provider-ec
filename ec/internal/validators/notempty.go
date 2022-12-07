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

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type notEmptyValidator struct{}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v notEmptyValidator) Description(ctx context.Context) string {
	return "Value must not be empty"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v notEmptyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and updating `resp` with diagnostics.
func (v notEmptyValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	if req.AttributeConfig.IsUnknown() || req.AttributeConfig.IsNull() || req.AttributeConfig.String() == "" {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			v.Description(ctx),
			"Value must be set and not empty",
		)
		return
	}
}

// NotEmpty returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is known.
//   - Is set.
//   - Is no empty string.
func NotEmpty() tfsdk.AttributeValidator {
	return notEmptyValidator{}
}
