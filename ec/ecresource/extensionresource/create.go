package extensionresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/extensionapi"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	var newState modelV0

	diags := request.Plan.Get(ctx, &newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	model, err := extensionapi.Create(
		extensionapi.CreateParams{
			API:         r.client,
			Name:        newState.Name.Value,
			Version:     newState.Version.Value,
			Type:        newState.ExtensionType.Value,
			Description: newState.Description.Value,
			DownloadURL: newState.DownloadURL.Value,
		},
	)
	if err != nil {
		response.Diagnostics.AddError(err.Error(), err.Error())
		return
	}

	newState.ID = types.String{Value: *model.ID}

	if !newState.FilePath.IsNull() && newState.FilePath.Value != "" {
		response.Diagnostics.Append(r.uploadExtension(newState)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	found, diags := r.read(newState.ID.Value, &newState)
	response.Diagnostics.Append(diags...)
	if !found {
		response.Diagnostics.AddError(
			"Failed to read deployment extension after create.",
			"Failed to read deployment extension after create.",
		)
		response.State.RemoveResource(ctx)
		return
	}
	if response.Diagnostics.HasError() {
		return
	}

	// Finally, set the state
	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}
