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
	"testing"

	resource_elasticsearch_project "github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_elasticsearch_project"
	resource_observability_project "github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_observability_project"
	resource_security_project "github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_security_project"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestOptionalMetadataForTagPatch_emptyPlanWithPriorTagsSendsNullRemovals(t *testing.T) {
	ctx := context.Background()
	planTags, _ := types.MapValue(types.StringType, map[string]attr.Value{})
	stateTags, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"k": basetypes.NewStringValue("v"),
	})
	om, diags := optionalMetadataForTagPatch(ctx, planTags, stateTags)
	require.False(t, diags.HasError())
	require.NotNil(t, om)
	require.Contains(t, om.Tags, "k")
	require.Nil(t, om.Tags["k"])
}

func TestOptionalMetadataForTagPatch_bothEmptyNoPatch(t *testing.T) {
	ctx := context.Background()
	planTags, _ := types.MapValue(types.StringType, map[string]attr.Value{})
	stateTags, _ := types.MapValue(types.StringType, map[string]attr.Value{})
	om, diags := optionalMetadataForTagPatch(ctx, planTags, stateTags)
	require.Nil(t, om)
	require.False(t, diags.HasError())
}

func TestOptionalMetadataForTagPatch_removedKeysAreJSONNull(t *testing.T) {
	ctx := context.Background()
	planTags, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"acc_team": basetypes.NewStringValue("platform"),
	})
	stateTags, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"acc_test": basetypes.NewStringValue("v2"),
	})
	om, diags := optionalMetadataForTagPatch(ctx, planTags, stateTags)
	require.False(t, diags.HasError())
	require.NotNil(t, om)
	require.NotNil(t, om.Tags["acc_team"])
	require.Equal(t, "platform", *om.Tags["acc_team"])
	require.Contains(t, om.Tags, "acc_test")
	require.Nil(t, om.Tags["acc_test"])
}

func TestPreserveElasticsearchMetadataSystemTags(t *testing.T) {
	ctx := context.Background()
	attrs := func(systemTags attr.Value) map[string]attr.Value {
		return map[string]attr.Value{
			"created_at":       basetypes.NewStringUnknown(),
			"created_by":       basetypes.NewStringUnknown(),
			"organization_id":  basetypes.NewStringUnknown(),
			"suspended_at":     basetypes.NewStringUnknown(),
			"suspended_reason": basetypes.NewStringUnknown(),
			"system_tags":      systemTags,
			"tags":             types.MapNull(types.StringType),
		}
	}
	attrTypes := resource_elasticsearch_project.MetadataValue{}.AttributeTypes(ctx)

	stateTags, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"managed-by": basetypes.NewStringValue("terraform"),
	})
	planTagsUnknown := types.MapUnknown(types.StringType)
	planTagsNull := types.MapNull(types.StringType)

	state := resource_elasticsearch_project.NewMetadataValueMust(attrTypes, attrs(stateTags))

	t.Run("copies system_tags from state when plan system_tags is unknown", func(t *testing.T) {
		plan := resource_elasticsearch_project.NewMetadataValueMust(attrTypes, attrs(planTagsUnknown))
		got := preserveElasticsearchMetadataSystemTags(plan, state)
		require.True(t, got.SystemTags.Equal(stateTags))
	})

	t.Run("copies system_tags from state when plan system_tags is null", func(t *testing.T) {
		plan := resource_elasticsearch_project.NewMetadataValueMust(attrTypes, attrs(planTagsNull))
		got := preserveElasticsearchMetadataSystemTags(plan, state)
		require.True(t, got.SystemTags.Equal(stateTags))
	})

	t.Run("keeps plan system_tags when set", func(t *testing.T) {
		planTags, _ := types.MapValue(types.StringType, map[string]attr.Value{
			"managed-by": basetypes.NewStringValue("manual"),
		})
		plan := resource_elasticsearch_project.NewMetadataValueMust(attrTypes, attrs(planTags))
		got := preserveElasticsearchMetadataSystemTags(plan, state)
		require.True(t, got.SystemTags.Equal(planTags))
	})

	t.Run("uses state when plan metadata is unknown", func(t *testing.T) {
		got := preserveElasticsearchMetadataSystemTags(resource_elasticsearch_project.NewMetadataValueUnknown(), state)
		require.True(t, got.Equal(state))
	})

	t.Run("keeps unknown plan when state is unknown", func(t *testing.T) {
		plan := resource_elasticsearch_project.NewMetadataValueMust(attrTypes, attrs(planTagsUnknown))
		got := preserveElasticsearchMetadataSystemTags(plan, resource_elasticsearch_project.NewMetadataValueUnknown())
		require.True(t, got.SystemTags.IsUnknown())
	})
}

func TestPreserveObservabilityMetadataSystemTags(t *testing.T) {
	ctx := context.Background()
	attrTypes := resource_observability_project.MetadataValue{}.AttributeTypes(ctx)
	stateTags, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"managed-by": basetypes.NewStringValue("terraform"),
	})
	state := resource_observability_project.NewMetadataValueMust(attrTypes, map[string]attr.Value{
		"created_at":       basetypes.NewStringUnknown(),
		"created_by":       basetypes.NewStringUnknown(),
		"organization_id":  basetypes.NewStringUnknown(),
		"suspended_at":     basetypes.NewStringUnknown(),
		"suspended_reason": basetypes.NewStringUnknown(),
		"system_tags":      stateTags,
		"tags":             types.MapNull(types.StringType),
	})
	plan := resource_observability_project.NewMetadataValueMust(attrTypes, map[string]attr.Value{
		"created_at":       basetypes.NewStringUnknown(),
		"created_by":       basetypes.NewStringUnknown(),
		"organization_id":  basetypes.NewStringUnknown(),
		"suspended_at":     basetypes.NewStringUnknown(),
		"suspended_reason": basetypes.NewStringUnknown(),
		"system_tags":      types.MapUnknown(types.StringType),
		"tags":             types.MapNull(types.StringType),
	})

	got := preserveObservabilityMetadataSystemTags(plan, state)
	require.True(t, got.SystemTags.Equal(stateTags))
}

func TestPreserveSecurityMetadataSystemTags(t *testing.T) {
	ctx := context.Background()
	attrTypes := resource_security_project.MetadataValue{}.AttributeTypes(ctx)
	stateTags, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"managed-by": basetypes.NewStringValue("terraform"),
	})
	state := resource_security_project.NewMetadataValueMust(attrTypes, map[string]attr.Value{
		"created_at":       basetypes.NewStringUnknown(),
		"created_by":       basetypes.NewStringUnknown(),
		"organization_id":  basetypes.NewStringUnknown(),
		"suspended_at":     basetypes.NewStringUnknown(),
		"suspended_reason": basetypes.NewStringUnknown(),
		"system_tags":      stateTags,
		"tags":             types.MapNull(types.StringType),
	})
	plan := resource_security_project.NewMetadataValueMust(attrTypes, map[string]attr.Value{
		"created_at":       basetypes.NewStringUnknown(),
		"created_by":       basetypes.NewStringUnknown(),
		"organization_id":  basetypes.NewStringUnknown(),
		"suspended_at":     basetypes.NewStringUnknown(),
		"suspended_reason": basetypes.NewStringUnknown(),
		"system_tags":      types.MapUnknown(types.StringType),
		"tags":             types.MapNull(types.StringType),
	})

	got := preserveSecurityMetadataSystemTags(plan, state)
	require.True(t, got.SystemTags.Equal(stateTags))
}
