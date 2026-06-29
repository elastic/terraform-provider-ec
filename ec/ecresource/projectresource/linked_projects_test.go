package projectresource

import (
	"testing"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestExpandLinkedProjectsForPatch(t *testing.T) {
	strValue := func(s string) attr.Value {
		return basetypes.NewStringValue(s)
	}
	toOptional := func(v attr.Value) *serverless.OptionalLinkedProject {
		return &serverless.OptionalLinkedProject{
			Type: serverless.ProjectType(v.(basetypes.StringValue).ValueString()),
		}
	}

	t.Run("adds new keys from plan", func(t *testing.T) {
		plan, _ := types.MapValue(types.StringType, map[string]attr.Value{
			"a": strValue("elasticsearch"),
		})
		patch := expandLinkedProjectsForPatch(plan, types.MapNull(types.StringType), toOptional)
		require.NotNil(t, patch)
		require.NotNil(t, patch.Projects)
		require.Contains(t, *patch.Projects, "a")
		require.Equal(t, serverless.ProjectType("elasticsearch"), (*patch.Projects)["a"].Type)
	})

	t.Run("keeps existing keys present in plan", func(t *testing.T) {
		plan, _ := types.MapValue(types.StringType, map[string]attr.Value{
			"a": strValue("elasticsearch"),
		})
		state, _ := types.MapValue(types.StringType, map[string]attr.Value{
			"a": strValue("elasticsearch"),
		})
		patch := expandLinkedProjectsForPatch(plan, state, toOptional)
		require.NotNil(t, patch)
		require.NotNil(t, patch.Projects)
		require.Contains(t, *patch.Projects, "a")
		require.Len(t, *patch.Projects, 1)
	})

	t.Run("emits nil for keys that have been removed from the plan", func(t *testing.T) {
		plan, _ := types.MapValue(types.StringType, map[string]attr.Value{
			"a": strValue("elasticsearch"),
		})
		state, _ := types.MapValue(types.StringType, map[string]attr.Value{
			"a": strValue("elasticsearch"),
			"b": strValue("observability"),
		})
		patch := expandLinkedProjectsForPatch(plan, state, toOptional)
		require.NotNil(t, patch)
		require.NotNil(t, patch.Projects)
		require.Contains(t, *patch.Projects, "a")
		require.Contains(t, *patch.Projects, "b")
		require.Equal(t, serverless.ProjectType("elasticsearch"), (*patch.Projects)["a"].Type)
		require.Nil(t, (*patch.Projects)["b"])
	})

	t.Run("returns nil when the planned map is empty and state is empty", func(t *testing.T) {
		patch := expandLinkedProjectsForPatch(types.MapNull(types.StringType), types.MapNull(types.StringType), toOptional)
		require.Nil(t, patch)
	})

	t.Run("emits nil for all keys when the planned map is empty", func(t *testing.T) {
		state, _ := types.MapValue(types.StringType, map[string]attr.Value{
			"a": strValue("elasticsearch"),
			"b": strValue("observability"),
		})
		patch := expandLinkedProjectsForPatch(types.MapNull(types.StringType), state, toOptional)
		require.NotNil(t, patch)
		require.NotNil(t, patch.Projects)
		require.Contains(t, *patch.Projects, "a")
		require.Contains(t, *patch.Projects, "b")
		require.Nil(t, (*patch.Projects)["a"])
		require.Nil(t, (*patch.Projects)["b"])
	})
}
