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

package utils

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	DataTiersVersion = semver.MustParse("7.10.0")
)

func UseNodeRoles(stateVersion, planVersion types.String) (bool, diag.Diagnostics) {

	useNodeRoles, err := CompatibleWithNodeRoles(planVersion.Value)

	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to determine whether to use node_roles", err.Error())
		return false, diags
	}

	convertLegacy, diags := LegacyToNodeRoles(stateVersion, planVersion)

	if diags.HasError() {
		return false, diags
	}

	return useNodeRoles && convertLegacy, nil
}

func CompatibleWithNodeRoles(version string) (bool, error) {
	deploymentVersion, err := semver.Parse(version)
	if err != nil {
		return false, fmt.Errorf("failed to parse Elasticsearch version: %w", err)
	}

	return deploymentVersion.GE(DataTiersVersion), nil
}

// LegacyToNodeRoles returns true when the legacy  "node_type_*" should be
// migrated over to node_roles. Which will be true when:
// * The version field doesn't change.
// * The version field changes but:
//   - The Elasticsearch.0.toplogy doesn't have any node_type_* set.
func LegacyToNodeRoles(stateVersion, planVersion types.String) (bool, diag.Diagnostics) {
	if stateVersion.Value == "" || stateVersion.Value == planVersion.Value {
		return true, nil
	}

	// If the previous version is empty, node_roles should be used.
	if stateVersion.Value == "" {
		return true, nil
	}

	var diags diag.Diagnostics
	oldV, err := semver.Parse(stateVersion.Value)
	if err != nil {
		diags.AddError("failed to parse previous Elasticsearch version", err.Error())
		return false, diags
	}
	newV, err := semver.Parse(planVersion.Value)
	if err != nil {
		diags.AddError("failed to parse new Elasticsearch version", err.Error())
		return false, diags
	}

	// if the version change moves from non-node_roles to one
	// that supports node roles, do not migrate on that step.
	if oldV.LT(DataTiersVersion) && newV.GE(DataTiersVersion) {
		return false, nil
	}

	return true, nil
}
