package extensionresource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/extensionapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/extensions"
)

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	if !resourceReady(r, &response.Diagnostics) {
		return
	}

	var state modelV0

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := extensionapi.Delete(extensionapi.DeleteParams{
		API:         r.client,
		ExtensionID: state.ID.Value,
	}); err != nil {
		if !alreadyDestroyed(err) {
			response.Diagnostics.AddError(err.Error(), err.Error())
		}
	}
}

func alreadyDestroyed(err error) bool {
	var extensionNotFound *extensions.DeleteExtensionNotFound
	return errors.As(err, &extensionNotFound)
}
