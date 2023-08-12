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

func VersionSupportsNodeRoles() validator.Set {
	return versionSupportsNodeRoles{}
}

var _ validator.Set = versionSupportsNodeRoles{}

type versionSupportsNodeRoles struct{}

func (r versionSupportsNodeRoles) Description(ctx context.Context) string {
	return "Validates the node_roles can only be defined if the stack version supports node roles."
}

func (r versionSupportsNodeRoles) MarkdownDescription(ctx context.Context) string {
	return "Validates the node_roles can only be defined if the stack version supports node roles."
}

// ValidateString should perform the validation.
func (v versionSupportsNodeRoles) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
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

	if utils.DataTiersVersion.GT(parsedVersion) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			fmt.Sprintf("[%s] not supported in the specified stack version", req.Path),
			fmt.Sprintf("The resources stack version [%s] does not support node_roles. Either convert your deployment resource to use node_types or use a stack version of at least [%s]", version, utils.DataTiersVersion),
		)
	}
}
