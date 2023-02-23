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

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = knownValidator{}

type knownValidator struct{}

func (v knownValidator) Description(ctx context.Context) string {
	return "Value must be known"
}

func (v knownValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v knownValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			v.Description(ctx),
			"Value must be known",
		)
		return
	}
}

// Known returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is known.
//
// Null (unconfigured) values are skipped.
func Known() knownValidator {
	return knownValidator{}
}
