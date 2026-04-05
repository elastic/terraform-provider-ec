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
	tags, ok := (*om)["tags"].(map[string]interface{})
	require.True(t, ok)
	require.Contains(t, tags, "k")
	require.Nil(t, tags["k"])
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
	tags, ok := (*om)["tags"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "platform", tags["acc_team"])
	require.Contains(t, tags, "acc_test")
	require.Nil(t, tags["acc_test"])
}
