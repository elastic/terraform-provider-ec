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

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

// emptyStringMap returns a known, empty map[string]string value, matching
// what metadataSystemTagsFromAPI / metadataTagsFromAPI produce when the API
// returns nil or empty tags.
func emptyStringMap() basetypes.MapValue {
	m, _ := types.MapValue(types.StringType, map[string]attr.Value{})
	return m
}

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

func TestMetadataSystemTagsFromAPI(t *testing.T) {
	ctx := context.Background()

	t.Run("returns an empty map when tags are nil", func(t *testing.T) {
		got, diags := metadataSystemTagsFromAPI(ctx, nil)
		require.False(t, diags.HasError())
		require.True(t, got.IsNull() || len(got.Elements()) == 0)
	})

	t.Run("returns an empty map when tags are empty", func(t *testing.T) {
		empty := serverless.ProjectSystemTags{}
		got, diags := metadataSystemTagsFromAPI(ctx, &empty)
		require.False(t, diags.HasError())
		require.True(t, got.IsNull() || len(got.Elements()) == 0)
	})

	t.Run("populates tags from the API response", func(t *testing.T) {
		tags := serverless.ProjectSystemTags{
			"managed-by": "terraform",
			"_foo":       "bar",
		}
		got, diags := metadataSystemTagsFromAPI(ctx, &tags)
		require.False(t, diags.HasError())
		require.False(t, got.IsNull())

		var m map[string]string
		diags2 := got.ElementsAs(ctx, &m, false)
		require.False(t, diags2.HasError())
		require.Equal(t, map[string]string{
			"managed-by": "terraform",
			"_foo":       "bar",
		}, m)
	})
}
