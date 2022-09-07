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

// NOTE! copied from terraform-provider-tls
package planmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// defaultValueAttributePlanModifier specifies a default value (attr.Value) for an attribute.
type defaultValueAttributePlanModifier struct {
	DefaultValue attr.Value
}

// DefaultValue is a helper to instantiate a defaultValueAttributePlanModifier.
func DefaultValue(v attr.Value) tfsdk.AttributePlanModifier {
	return &defaultValueAttributePlanModifier{v}
}

var _ tfsdk.AttributePlanModifier = (*defaultValueAttributePlanModifier)(nil)

func (m *defaultValueAttributePlanModifier) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m *defaultValueAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Sets the default value %q (%s) if the attribute is not set", m.DefaultValue, m.DefaultValue.Type(ctx))
}

func (m *defaultValueAttributePlanModifier) Modify(_ context.Context, req tfsdk.ModifyAttributePlanRequest, res *tfsdk.ModifyAttributePlanResponse) {
	// If the attribute configuration is not null, we are done here
	if !req.AttributeConfig.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan m in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.AttributePlan.IsUnknown() && !req.AttributePlan.IsNull() {
		return
	}

	res.AttributePlan = m.DefaultValue
}
