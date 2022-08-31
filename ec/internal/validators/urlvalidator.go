package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
	"net/url"
	"strings"
)

type isURLWithSchemeValidator struct {
	ValidSchemes []string
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v isURLWithSchemeValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Value must be a valid URL with scheme (%s)", strings.Join(v.ValidSchemes, ", "))
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v isURLWithSchemeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and updating `resp` with diagnostics.
func (v isURLWithSchemeValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// types.String must be the attr.Value produced by the attr.Type in the schema for this attribute
	// for generic validators, use
	// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#ConvertValue
	// to convert into a known type.
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if str.Unknown || str.Null {
		return
	}

	if str.Value == "" {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			v.Description(ctx),
			fmt.Sprintf("URL must not be empty, got %v.", str),
		)
		return
	}

	u, err := url.Parse(str.Value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			v.Description(ctx),
			fmt.Sprintf("URL is invalid, got %v: %+v", str.Value, err),
		)
		return
	}

	if u.Host == "" {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			v.Description(ctx),
			fmt.Sprintf("URL is missing host, got %v", str.Value),
		)
		return
	}

	if !slices.Contains(v.ValidSchemes, u.Scheme) {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			v.Description(ctx),
			fmt.Sprintf("URL is expected to have a valid scheme, got %v (%v)", u.Scheme, str.Value),
		)
	}
	return
}

func IsURLWithSchemeValidator(validSchemes []string) tfsdk.AttributeValidator {
	return isURLWithSchemeValidator{}
}
