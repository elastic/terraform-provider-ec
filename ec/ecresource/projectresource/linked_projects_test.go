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

	// mapStr builds a typed map from id -> type string. A nil input produces a
	// null map (the framework representation of "absent").
	mapStr := func(m map[string]string) basetypes.MapValue {
		if m == nil {
			return types.MapNull(types.StringType)
		}
		elems := make(map[string]attr.Value, len(m))
		for k, v := range m {
			elems[k] = strValue(v)
		}
		out, _ := types.MapValue(types.StringType, elems)
		return out
	}

	tests := []struct {
		name         string
		plan         map[string]string
		state        map[string]string
		wantNilPatch bool
		wantTyped    map[string]string // keys expected present with the given type
		wantUnlinked []string          // keys expected present as nil
	}{
		{
			name:      "adds new keys from plan",
			plan:      map[string]string{"a": "elasticsearch"},
			state:     nil,
			wantTyped: map[string]string{"a": "elasticsearch"},
		},
		{
			name:      "keeps existing keys present in plan",
			plan:      map[string]string{"a": "elasticsearch"},
			state:     map[string]string{"a": "elasticsearch"},
			wantTyped: map[string]string{"a": "elasticsearch"},
		},
		{
			name:         "emits nil for keys removed from the plan",
			plan:         map[string]string{"a": "elasticsearch"},
			state:        map[string]string{"a": "elasticsearch", "b": "observability"},
			wantTyped:    map[string]string{"a": "elasticsearch"},
			wantUnlinked: []string{"b"},
		},
		{
			name:         "emits nil for all keys when the whole plan is removed",
			plan:         nil,
			state:        map[string]string{"a": "elasticsearch", "b": "observability"},
			wantUnlinked: []string{"a", "b"},
		},
		{
			name:         "returns nil when plan and state are empty",
			plan:         nil,
			state:        nil,
			wantNilPatch: true,
		},
		{
			name:      "ignores unknown state",
			plan:      map[string]string{"a": "elasticsearch"},
			state:     nil,
			wantTyped: map[string]string{"a": "elasticsearch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch := expandLinkedProjectsForPatch(mapStr(tt.plan), mapStr(tt.state), toOptional)

			if tt.wantNilPatch {
				require.Nil(t, patch)
				return
			}

			require.NotNil(t, patch)
			require.NotNil(t, patch.Projects)

			// All typed keys present with the expected type.
			for id, typ := range tt.wantTyped {
				v, ok := (*patch.Projects)[id]
				require.True(t, ok, "expected key %q to be present", id)
				require.NotNil(t, v, "expected key %q to be non-nil (typed)", id)
				require.Equal(t, serverless.ProjectType(typ), v.Type)
			}

			// All unlinked keys present as explicit nil.
			for _, id := range tt.wantUnlinked {
				v, ok := (*patch.Projects)[id]
				require.True(t, ok, "expected unlinked key %q to be present (as nil)", id)
				require.Nil(t, v, "expected key %q to be nil (unlink)", id)
			}

			// No other keys.
			require.Len(t, *patch.Projects, len(tt.wantTyped)+len(tt.wantUnlinked))
		})
	}
}
