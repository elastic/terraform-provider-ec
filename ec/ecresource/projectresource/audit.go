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
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// monitoring is the planned or stored monitoring block, independent of which
// project resource owns the Terraform types. It mirrors the schema: nil is an
// omitted block, a block with nil logging or nil audit is a shell, and a
// non-nil audit is configured.
type monitoring struct {
	Logging *logging
}

type logging struct {
	Audit *audit
}

type audit struct {
	Enabled     bool
	ProjectID   string
	ProjectType string
	// IgnoreFilterIDs is nil for a null set and empty for an empty set.
	IgnoreFilterIDs []string
}

func (m *monitoring) audit() *audit {
	if m == nil || m.Logging == nil {
		return nil
	}
	return m.Logging.Audit
}

func (m *monitoring) shell() bool {
	return m != nil && m.audit() == nil
}

func (a *audit) hasNonEmptyFilters() bool {
	return a != nil && len(a.IgnoreFilterIDs) > 0
}

// expandMonitoringForCreate builds the create request's monitoring field.
// Audit unset omits monitoring. A null or empty ignore-filter set omits
// ignore_filters. Destination status is not part of the create body.
func expandMonitoringForCreate(in *monitoring) *serverless.MonitoringCreateRequest {
	a := in.audit()
	if a == nil {
		return nil
	}
	projectType := serverless.ProjectType(a.ProjectType)
	return &serverless.MonitoringCreateRequest{
		Logging: &serverless.LoggingCreateRequest{
			Audit: &serverless.LoggingCategoryCreateRequest{
				Enabled: &a.Enabled,
				Destination: serverless.LoggingDestinationCreateRequest{
					ProjectId:   a.ProjectID,
					ProjectType: &projectType,
				},
				IgnoreFilters: expandIgnoreFilterIDs(a.IgnoreFilterIDs),
			},
		},
	}
}

// expandIgnoreFilterIDs converts ignore-filter id strings to the API's {id}
// objects. Nil and an empty slice both omit the field.
func expandIgnoreFilterIDs(ids []string) *[]serverless.IgnoreFilterRef {
	if len(ids) == 0 {
		return nil
	}
	// The Terraform attribute is a set. Send a stable order so the request
	// does not change when set iteration order changes.
	sorted := append([]string(nil), ids...)
	sort.Strings(sorted)
	refs := make([]serverless.IgnoreFilterRef, len(sorted))
	for i, id := range sorted {
		refs[i] = serverless.IgnoreFilterRef{Id: id}
	}
	return &refs
}

// flattenIgnoreFilterIDs converts the API's {id} objects back to id strings.
// A nil slice stays nil.
func flattenIgnoreFilterIDs(filters *[]serverless.IgnoreFilterRef) []string {
	if filters == nil {
		return nil
	}
	ids := make([]string, len(*filters))
	for i, filter := range *filters {
		ids[i] = filter.Id
	}
	return ids
}

// loggingObjectModel, auditObjectModel, and destinationObjectModel decode a
// resource's monitoring.logging object with the framework's reflection, so the
// shared conversion does not depend on a resource's generated value types. Each
// struct lists every attribute of its object, as reflection requires.
type loggingObjectModel struct {
	Audit types.Object `tfsdk:"audit"`
}

type auditObjectModel struct {
	Enabled         types.Bool   `tfsdk:"enabled"`
	Destination     types.Object `tfsdk:"destination"`
	IgnoreFilterIDs types.Set    `tfsdk:"ignore_filter_ids"`
}

type destinationObjectModel struct {
	ProjectID   types.String `tfsdk:"project_id"`
	ProjectType types.String `tfsdk:"project_type"`
	Status      types.String `tfsdk:"status"`
}

// monitoringFrom converts a resource's monitoring value. monitoringNull is the
// top-level block's IsNull; loggingObj is its logging attribute.
func monitoringFrom(ctx context.Context, monitoringNull bool, loggingObj basetypes.ObjectValue) (*monitoring, diag.Diagnostics) {
	var diags diag.Diagnostics
	if monitoringNull {
		return nil, diags
	}
	m := &monitoring{}
	if loggingObj.IsNull() {
		return m, diags
	}
	m.Logging = &logging{}
	if loggingObj.IsUnknown() {
		return m, diags
	}

	var loggingModel loggingObjectModel
	diags.Append(loggingObj.As(ctx, &loggingModel, basetypes.ObjectAsOptions{})...)
	if diags.HasError() || loggingModel.Audit.IsNull() {
		return m, diags
	}

	// An unknown audit object has not been filled in yet; decode it as empty so
	// the fields below fall back to their defaults.
	var auditModel auditObjectModel
	diags.Append(loggingModel.Audit.As(ctx, &auditModel, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return m, diags
	}

	// The schema default is true. A null or unknown plan value has not been
	// filled yet, so send the default rather than false.
	a := &audit{Enabled: true}
	if util.IsKnown(auditModel.Enabled) {
		a.Enabled = auditModel.Enabled.ValueBool()
	}
	if util.IsKnown(auditModel.Destination) {
		var dest destinationObjectModel
		diags.Append(auditModel.Destination.As(ctx, &dest, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return m, diags
		}
		if util.IsKnown(dest.ProjectID) {
			a.ProjectID = dest.ProjectID.ValueString()
		}
		if util.IsKnown(dest.ProjectType) {
			a.ProjectType = dest.ProjectType.ValueString()
		}
	}
	if util.IsKnown(auditModel.IgnoreFilterIDs) {
		diags.Append(auditModel.IgnoreFilterIDs.ElementsAs(ctx, &a.IgnoreFilterIDs, false)...)
		if diags.HasError() {
			return m, diags
		}
	}
	m.Logging.Audit = a
	return m, diags
}

// firstError returns the first error diagnostic as an error, for callers that
// return error rather than diagnostics.
func firstError(diags diag.Diagnostics) error {
	for _, d := range diags {
		if d.Severity() == diag.SeverityError {
			return fmt.Errorf("%s: %s", d.Summary(), d.Detail())
		}
	}
	return nil
}

// auditPatch is the monitoring field of a patch. raw is set when the body
// must contain JSON null, which the generated omitempty pointers cannot encode.
type auditPatch struct {
	raw   json.RawMessage
	typed *serverless.OptionalMonitoring
}

func (p auditPatch) useRaw() bool {
	return len(p.raw) > 0
}

func auditPatchFor(plan, state *monitoring) (auditPatch, error) {
	planAudit, stateAudit := plan.audit(), state.audit()
	if planAudit != nil {
		if stateAudit.hasNonEmptyFilters() && !planAudit.hasNonEmptyFilters() {
			raw, err := json.Marshal(map[string]any{
				"logging": map[string]any{
					"audit": map[string]any{
						"enabled": planAudit.Enabled,
						"destination": map[string]any{
							"project_id":   planAudit.ProjectID,
							"project_type": planAudit.ProjectType,
						},
						"ignore_filters": nil,
					},
				},
			})
			return auditPatch{raw: raw}, err
		}
		return auditPatch{typed: optionalMonitoring(planAudit)}, nil
	}
	if stateAudit != nil {
		raw, err := json.Marshal(map[string]any{
			"logging": map[string]any{"audit": nil},
		})
		return auditPatch{raw: raw}, err
	}
	return auditPatch{}, nil
}

func optionalMonitoring(a *audit) *serverless.OptionalMonitoring {
	enabled := a.Enabled
	projectType := serverless.ProjectType(a.ProjectType)
	return &serverless.OptionalMonitoring{
		Logging: &serverless.OptionalLogging{
			Audit: &serverless.OptionalLoggingCategory{
				Enabled: &enabled,
				Destination: &serverless.LoggingDestinationCreateRequest{
					ProjectId:   a.ProjectID,
					ProjectType: &projectType,
				},
				IgnoreFilters: expandIgnoreFilterIDs(a.IgnoreFilterIDs),
			},
		},
	}
}

// mergeMonitoringField marshals body, then sets its monitoring field to the
// already-encoded JSON value. Callers use this when that value contains null.
func mergeMonitoringField(body any, monitoring json.RawMessage) ([]byte, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	doc := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	doc["monitoring"] = monitoring
	return json.Marshal(doc)
}

// flattenMonitoring returns the monitoring object Read should store. It is
// null when the block was omitted or the API no longer has audit, the prior
// shell when there is no audit to read into it, and the API audit otherwise.
func flattenMonitoring(monitoringTypes map[string]attr.Type, prior *monitoring, api *serverless.Monitoring) (basetypes.ObjectValue, diag.Diagnostics) {
	null := types.ObjectNull(monitoringTypes)
	apiAudit := apiAuditCategory(api)
	if apiAudit == nil && !prior.shell() {
		return null, nil
	}

	loggingTypes, diags := objectAttrTypes(monitoringTypes, "logging")
	if diags.HasError() {
		return null, diags
	}
	auditTypes, auditDiags := objectAttrTypes(loggingTypes, "audit")
	diags.Append(auditDiags...)
	if diags.HasError() {
		return null, diags
	}

	var auditObj attr.Value = types.ObjectNull(auditTypes)
	if apiAudit != nil {
		var auditValDiags diag.Diagnostics
		auditObj, auditValDiags = flattenAudit(auditTypes, prior.audit(), apiAudit)
		diags.Append(auditValDiags...)
		if diags.HasError() {
			return null, diags
		}
	}

	var loggingObj attr.Value = types.ObjectNull(loggingTypes)
	if apiAudit != nil || prior.Logging != nil {
		var loggingDiags diag.Diagnostics
		loggingObj, loggingDiags = types.ObjectValue(loggingTypes, map[string]attr.Value{"audit": auditObj})
		diags.Append(loggingDiags...)
		if diags.HasError() {
			return null, diags
		}
	}

	obj, objDiags := types.ObjectValue(monitoringTypes, map[string]attr.Value{"logging": loggingObj})
	diags.Append(objDiags...)
	if diags.HasError() {
		return null, diags
	}
	return obj, diags
}

func flattenAudit(auditTypes map[string]attr.Type, prior *audit, apiAudit *serverless.LoggingCategory) (basetypes.ObjectValue, diag.Diagnostics) {
	null := types.ObjectNull(auditTypes)
	destTypes, diags := objectAttrTypes(auditTypes, "destination")
	if diags.HasError() {
		return null, diags
	}

	filterVal, filterDiags := filterSet(readFilterIDs(prior, apiAudit.IgnoreFilters))
	diags.Append(filterDiags...)
	if diags.HasError() {
		return null, diags
	}

	dest, destValDiags := types.ObjectValue(destTypes, map[string]attr.Value{
		"project_id":   types.StringValue(apiAudit.Destination.ProjectId),
		"project_type": types.StringValue(string(apiAudit.Destination.ProjectType)),
		"status":       types.StringValue(string(apiAudit.Destination.Status)),
	})
	diags.Append(destValDiags...)
	if diags.HasError() {
		return null, diags
	}
	obj, auditValDiags := types.ObjectValue(auditTypes, map[string]attr.Value{
		"destination":       dest,
		"enabled":           types.BoolValue(apiAudit.Enabled),
		"ignore_filter_ids": filterVal,
	})
	diags.Append(auditValDiags...)
	if diags.HasError() {
		return null, diags
	}
	return obj, diags
}

func apiAuditCategory(api *serverless.Monitoring) *serverless.LoggingCategory {
	if api == nil || api.Logging == nil {
		return nil
	}
	return api.Logging.Audit
}

// readFilterIDs picks the ignore_filter_ids to store. API ids win. With no API
// ids, a prior empty set stays empty and anything else is null.
func readFilterIDs(prior *audit, apiFilters *[]serverless.IgnoreFilterRef) []string {
	flat := flattenIgnoreFilterIDs(apiFilters)
	if len(flat) > 0 {
		return flat
	}
	if prior != nil && prior.IgnoreFilterIDs != nil && len(prior.IgnoreFilterIDs) == 0 {
		return []string{}
	}
	return nil
}

func filterSet(ids []string) (basetypes.SetValue, diag.Diagnostics) {
	if ids == nil {
		return types.SetNull(types.StringType), nil
	}
	elems := make([]attr.Value, len(ids))
	for i, id := range ids {
		elems[i] = types.StringValue(id)
	}
	return types.SetValue(types.StringType, elems)
}

func objectAttrTypes(parent map[string]attr.Type, name string) (map[string]attr.Type, diag.Diagnostics) {
	typed, ok := parent[name].(attr.TypeWithAttributeTypes)
	if !ok || typed == nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid monitoring schema",
				fmt.Sprintf("attribute %q is not an object", name),
			),
		}
	}
	return typed.AttributeTypes(), nil
}
