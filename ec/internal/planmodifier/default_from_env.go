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

package planmodifier

import (
	"context"
	"fmt"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// defaultFromEnvAttributePlanModifier specifies a default value (attr.Value) for an attribute.
type defaultFromEnvAttributePlanModifier struct {
	EnvKeys []string
}

// DefaultFromEnv is a helper to instantiate a defaultFromEnvAttributePlanModifier.
func DefaultFromEnv(envKeys []string) tfsdk.AttributePlanModifier {
	return &defaultFromEnvAttributePlanModifier{envKeys}
}

var _ tfsdk.AttributePlanModifier = (*defaultFromEnvAttributePlanModifier)(nil)

func (m *defaultFromEnvAttributePlanModifier) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m *defaultFromEnvAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Sets the default value from an environment variable (%v) if the attribute is not set", m.EnvKeys)
}

func (m *defaultFromEnvAttributePlanModifier) Modify(_ context.Context, req tfsdk.ModifyAttributePlanRequest, res *tfsdk.ModifyAttributePlanResponse) {
	// If the attribute configuration is not null, we are done here
	if !req.AttributeConfig.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan m in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.AttributePlan.IsUnknown() && !req.AttributePlan.IsNull() {
		return
	}

	res.AttributePlan = types.String{Value: util.MultiGetenv(m.EnvKeys, "")}
}
