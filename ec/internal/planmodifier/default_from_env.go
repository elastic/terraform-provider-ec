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
