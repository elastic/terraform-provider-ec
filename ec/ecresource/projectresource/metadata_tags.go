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
	"maps"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// patchMetadataSchema relaxes metadata.tags and adds plan modifiers so computed metadata
// attributes (created_at, etc.) do not become unknown when only tags are set in config.
func patchMetadataSchema(resp *resource.SchemaResponse) {
	patchMetadataTagsSchema(resp)

	metaAttr, ok := resp.Schema.Attributes["metadata"].(schema.SingleNestedAttribute)
	if !ok {
		return
	}
	for _, key := range []string{"created_at", "created_by", "organization_id", "suspended_at", "suspended_reason"} {
		sa, ok := metaAttr.Attributes[key].(schema.StringAttribute)
		if !ok {
			continue
		}
		sa.PlanModifiers = append(sa.PlanModifiers, stringplanmodifier.UseStateForUnknown())
		metaAttr.Attributes[key] = sa
	}
	resp.Schema.Attributes["metadata"] = metaAttr
}

// patchMetadataTagsSchema relaxes generated metadata.tags (required + min size 1) so
// API responses and configurations without tags remain valid; tag count is still capped at 64.
func patchMetadataTagsSchema(resp *resource.SchemaResponse) {
	metaAttr, ok := resp.Schema.Attributes["metadata"].(schema.SingleNestedAttribute)
	if !ok {
		return
	}
	tagsAttr, ok := metaAttr.Attributes["tags"].(schema.MapAttribute)
	if !ok {
		return
	}
	tagsAttr.Required = false
	tagsAttr.Optional = true
	tagsAttr.Computed = true
	tagsAttr.Validators = []validator.Map{
		mapvalidator.SizeBetween(0, 64),
	}
	metaAttr.Attributes["tags"] = tagsAttr
	resp.Schema.Attributes["metadata"] = metaAttr
}

func metadataTagsFromAPI(ctx context.Context, tags *serverless.ProjectTags) (basetypes.MapValue, diag.Diagnostics) {
	if tags == nil || len(*tags) == 0 {
		m, d := types.MapValue(types.StringType, map[string]attr.Value{})
		return m, d
	}
	elems := make(map[string]attr.Value, len(*tags))
	for k, v := range *tags {
		elems[k] = basetypes.NewStringValue(string(v))
	}
	return types.MapValue(types.StringType, elems)
}

func projectMetadataRequestFromTFMetadata(ctx context.Context, tags basetypes.MapValue) (*serverless.ProjectMetadataRequest, diag.Diagnostics) {
	if !util.IsKnown(tags) || tags.IsNull() {
		return nil, nil
	}
	var tagMap map[string]string
	diags := tags.ElementsAs(ctx, &tagMap, false)
	if diags.HasError() {
		return nil, diags
	}
	if len(tagMap) == 0 {
		return nil, diags
	}
	out := make(serverless.ProjectTags, len(tagMap))
	for k, v := range tagMap {
		out[k] = serverless.ProjectTagValue(v)
	}
	return &serverless.ProjectMetadataRequest{Tags: out}, diags
}

// tagMapFromTF returns an empty map for null/unknown tags.
func tagMapFromTF(ctx context.Context, tags basetypes.MapValue) (map[string]string, diag.Diagnostics) {
	if !util.IsKnown(tags) || tags.IsNull() {
		return map[string]string{}, nil
	}
	var m map[string]string
	diags := tags.ElementsAs(ctx, &m, false)
	if m == nil {
		m = map[string]string{}
	}
	return m, diags
}

// optionalMetadataForTagPatch builds patch metadata when plan tags differ from state tags.
// Tag updates use JSON Merge Patch semantics: keys removed in config are sent as JSON null so
// the API drops them; an empty planned map while state still has tags sends null for every
// prior key (full clear), if the API accepts that shape.
func optionalMetadataForTagPatch(ctx context.Context, planTags, stateTags basetypes.MapValue) (*serverless.OptionalMetadata, diag.Diagnostics) {
	if !util.IsKnown(planTags) || planTags.IsNull() {
		return nil, nil
	}
	planMap, diags := tagMapFromTF(ctx, planTags)
	if diags.HasError() {
		return nil, diags
	}
	stateMap, d2 := tagMapFromTF(ctx, stateTags)
	diags = append(diags, d2...)
	if d2.HasError() {
		return nil, diags
	}
	if maps.Equal(planMap, stateMap) {
		return nil, diags
	}
	var wrapped map[string]interface{}
	if len(planMap) == 0 && len(stateMap) > 0 {
		wrapped = make(map[string]interface{}, len(stateMap))
		for k := range stateMap {
			wrapped[k] = nil
		}
	} else {
		wrapped = make(map[string]interface{}, len(planMap)+len(stateMap))
		for k, v := range planMap {
			wrapped[k] = v
		}
		for k := range stateMap {
			if _, ok := planMap[k]; !ok {
				wrapped[k] = nil
			}
		}
	}
	om := serverless.OptionalMetadata{"tags": wrapped}
	return &om, diags
}
