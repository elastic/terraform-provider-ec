package v2

import (
	"context"

	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type useNullIfNotAPM struct{}

var _ planmodifier.String = useNullIfNotAPM{}

func (m useNullIfNotAPM) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m useNullIfNotAPM) MarkdownDescription(ctx context.Context) string {
	return "Sets the plan value to null if there is no apm or integrations_server resource"
}

func (m useNullIfNotAPM) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if req.ConfigValue.IsUnknown() {
		return
	}

	if !req.PlanValue.IsUnknown() {
		return
	}

	hasAPM, diags := planmodifiers.HasAttribute(ctx, path.Root("apm"), req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if hasAPM {
		return
	}

	hasIntegrationsServer, diags := planmodifiers.HasAttribute(ctx, path.Root("integrations_server"), req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if hasIntegrationsServer {
		return
	}

	resp.PlanValue = types.StringNull()
}
