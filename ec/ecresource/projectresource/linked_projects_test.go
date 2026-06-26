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

	t.Run("emits nil for removed keys", func(t *testing.T) {
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
		require.NotNil(t, (*patch.Projects)["a"])
		require.Nil(t, (*patch.Projects)["b"])
	})

	t.Run("unlinks everything when the whole block is removed", func(t *testing.T) {
		state, _ := types.MapValue(types.StringType, map[string]attr.Value{
			"a": strValue("elasticsearch"),
			"b": strValue("observability"),
		})
		patch := expandLinkedProjectsForPatch(types.MapNull(types.StringType), state, toOptional)
		require.NotNil(t, patch)
		require.NotNil(t, patch.Projects)
		require.Len(t, *patch.Projects, 2)
		require.Nil(t, (*patch.Projects)["a"])
		require.Nil(t, (*patch.Projects)["b"])
	})

	t.Run("returns nil when plan and state are empty", func(t *testing.T) {
		patch := expandLinkedProjectsForPatch(types.MapNull(types.StringType), types.MapNull(types.StringType), toOptional)
		require.Nil(t, patch)
	})

	t.Run("returns nil when state is empty and plan has no links", func(t *testing.T) {
		patch := expandLinkedProjectsForPatch(types.MapNull(types.StringType), types.MapNull(types.StringType), toOptional)
		require.Nil(t, patch)
	})
}
