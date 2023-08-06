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
	"fmt"

	"github.com/blang/semver"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = versionSupportsNodeTypes{}

func VersionSupportsNodeTypes() validator.String {
	return versionSupportsNodeTypes{}
}

type versionSupportsNodeTypes struct{}

func (r versionSupportsNodeTypes) Description(ctx context.Context) string {
	return "Validates the node_types can only be defined if the stack version supports node types."
}

func (r versionSupportsNodeTypes) MarkdownDescription(ctx context.Context) string {
	return "Validates the node_types can only be defined if the stack version supports node types."
}

// ValidateString should perform the validation.
func (v versionSupportsNodeTypes) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var version string
	resp.Diagnostics = req.Config.GetAttribute(ctx, path.Root("version"), &version)
	if resp.Diagnostics.HasError() {
		return
	}

	parsedVersion, err := semver.Parse(version)
	if err != nil {
		// Ignore this error, it's validated as part of the version schema definition
		return
	}

	if utils.MinVersionWithoutNodeTypes.LTE(parsedVersion) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			fmt.Sprintf("[%s] not supported in the specified stack version", req.Path),
			fmt.Sprintf("The resources stack version [%s] does not support node_types. Either convert your deployment resource to use node_roles or use a stack version less than [%s]", version, utils.MinVersionWithoutNodeTypes),
		)
	}
}
