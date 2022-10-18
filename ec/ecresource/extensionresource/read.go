package extensionresource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/extensionapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/extensions"
	"github.com/elastic/cloud-sdk-go/pkg/models"
)

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	var newState modelV0

	diags := request.State.Get(ctx, &newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(newState.ID.Value, &newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	if !found {
		response.State.RemoveResource(ctx)
		return
	}

	// Finally, set the state
	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

func (r *Resource) read(id string, state *modelV0) (found bool, diags diag.Diagnostics) {
	res, err := extensionapi.Get(extensionapi.GetParams{
		API:         r.client,
		ExtensionID: id,
	})
	if err != nil {
		if extensionNotFound(err) {
			return false, diags
		}
		diags.AddError("failed reading extension", err.Error())
		return true, diags
	}

	modelToState(res, state)
	return true, diags
}

func extensionNotFound(err error) bool {
	// We're using the As() call since we do not care about the error value
	// but do care about the error's contents type since it's an implicit 404.
	var extensionNotFound *extensions.GetExtensionNotFound
	return errors.As(err, &extensionNotFound)
}

func modelToState(model *models.Extension, state *modelV0) {
	if model.Name != nil {
		state.Name = types.String{Value: *model.Name}
	} else {
		state.Name = types.String{Null: true}
	}

	if model.Version != nil {
		state.Version = types.String{Value: *model.Version}
	} else {
		state.Version = types.String{Null: true}
	}

	if model.ExtensionType != nil {
		state.ExtensionType = types.String{Value: *model.ExtensionType}
	} else {
		state.ExtensionType = types.String{Null: true}
	}

	state.Description = types.String{Value: model.Description}

	if model.URL != nil {
		state.URL = types.String{Value: *model.URL}
	} else {
		state.URL = types.String{Null: true}
	}

	state.DownloadURL = types.String{Value: model.DownloadURL}

	if metadata := model.FileMetadata; metadata != nil {
		state.LastModified = types.String{Value: metadata.LastModifiedDate.String()}
		state.Size = types.Int64{Value: metadata.Size}
	} else {
		state.LastModified = types.String{Null: true}
		state.Size = types.Int64{Null: true}
	}
}
