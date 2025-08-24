package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func UseStateIfNotNullForUnknown() useStateIfNotNullForKnown {
	return useStateIfNotNullForKnown{}
}

// useStateIfNotNullForKnown implements the plan modifier.
type useStateIfNotNullForKnown struct{}

// Description returns a human-readable description of the plan modifier.
func (m useStateIfNotNullForKnown) Description(_ context.Context) string {
	return "Set the plan value to unknown if the state value is not known."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m useStateIfNotNullForKnown) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// PlanModifyString implements the plan modification logic.
func (m useStateIfNotNullForKnown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !m.shouldUseState(req.State, req.PlanValue, req.ConfigValue, req.StateValue) {
		return
	}

	resp.PlanValue = req.StateValue
}

// PlanModifyBool implements the plan modification logic.
func (m useStateIfNotNullForKnown) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !m.shouldUseState(req.State, req.PlanValue, req.ConfigValue, req.StateValue) {
		return
	}

	resp.PlanValue = req.StateValue
}

func (m useStateIfNotNullForKnown) shouldUseState(state tfsdk.State, planValue, configValue, stateValue attr.Value) bool {
	// Do nothing if there is no state (resource is being created).
	if state.Raw.IsNull() {
		return false
	}

	// Do nothing if there is a known planned value.
	if !planValue.IsUnknown() {
		return false
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if configValue.IsUnknown() {
		return false
	}

	// Do nothing if the state value is null
	if stateValue.IsNull() {
		return false
	}

	return true
}
