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

package projectresource

import (
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// expandLinkedProjectsForPatch builds an OptionalLinkConfiguration patch body from
// the planned and prior-state linked project maps. Keys present in the plan are
// added/updated. Keys that existed in state but are no longer in the plan are
// emitted as an explicit null, because the serverless PATCH API treats an
// omitted key as "no change" rather than "unlink" — omission would silently
// leave the link intact.
func expandLinkedProjectsForPatch(
	planProjects, stateProjects basetypes.MapValue,
	toOptional func(attr.Value) *serverless.OptionalLinkedProject,
) *serverless.OptionalLinkConfiguration {
	planElems := map[string]attr.Value{}
	if util.IsKnown(planProjects) && !planProjects.IsNull() {
		planElems = planProjects.Elements()
	}

	stateElems := map[string]attr.Value{}
	if util.IsKnown(stateProjects) && !stateProjects.IsNull() {
		stateElems = stateProjects.Elements()
	}

	if len(planElems) == 0 && len(stateElems) == 0 {
		return nil
	}

	projects := make(map[string]*serverless.OptionalLinkedProject, max(len(planElems), len(stateElems)))

	// Add/update keys present in the plan.
	for id, v := range planElems {
		projects[id] = toOptional(v)
	}

	// Unlink keys that existed in state but are no longer in the plan by emitting
	// an explicit null. Omitting the key would leave the link intact.
	for id := range stateElems {
		if _, ok := planElems[id]; !ok {
			projects[id] = nil
		}
	}

	return &serverless.OptionalLinkConfiguration{Projects: &projects}
}
