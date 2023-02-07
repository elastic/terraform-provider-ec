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
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/exp/slices"
)

type isURLWithSchemeValidator struct {
	ValidSchemes []string
}

func (v isURLWithSchemeValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Value must be a valid URL with scheme (%s)", strings.Join(v.ValidSchemes, ", "))
}

func (v isURLWithSchemeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isURLWithSchemeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if req.ConfigValue.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			v.Description(ctx),
			fmt.Sprintf("URL must not be empty, got %v.", req.ConfigValue.ValueString()),
		)
		return
	}

	u, err := url.Parse(req.ConfigValue.ValueString())

	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			v.Description(ctx),
			fmt.Sprintf("URL is invalid, got %v: %+v", req.ConfigValue.ValueString(), err),
		)
		return
	}

	if u.Host == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			v.Description(ctx),
			fmt.Sprintf("URL is missing host, got %v", req.ConfigValue.ValueString()),
		)
		return
	}

	if !slices.Contains(v.ValidSchemes, u.Scheme) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			v.Description(ctx),
			fmt.Sprintf("URL is expected to have a valid scheme (one of '%v'), got %v (%v)", v.ValidSchemes, u.Scheme, req.ConfigValue.ValueString()),
		)
	}
}

func IsURLWithSchemeValidator(validSchemes []string) validator.String {
	return isURLWithSchemeValidator{ValidSchemes: validSchemes}
}
