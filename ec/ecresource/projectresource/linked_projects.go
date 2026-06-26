package projectresource

import (
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// expandLinkedProjectsForPatch builds an OptionalLinkConfiguration patch body from
// the planned and prior-state linked project maps. Keys present in the plan are
// added/updated; keys that existed in state but have been removed from the plan
// are emitted as null so the API unlinks them.
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

	// Emit null for keys that existed in state but are no longer in the plan.
	for id := range stateElems {
		if _, ok := planElems[id]; !ok {
			projects[id] = nil
		}
	}

	return &serverless.OptionalLinkConfiguration{Projects: &projects}
}
