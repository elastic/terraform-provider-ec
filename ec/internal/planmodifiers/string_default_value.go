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
package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.String = stringDefaultValue{}

type stringDefaultValue struct {
	value string
}

func StringDefaultValue(v string) stringDefaultValue {
	return stringDefaultValue{v}
}

func (m stringDefaultValue) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m stringDefaultValue) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Sets the default value %v if the attribute is not set", m.value)
}

func (m stringDefaultValue) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = types.StringValue(m.value)
}
