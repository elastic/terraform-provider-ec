package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestElasticsearchConfig_IsEmpty(t *testing.T) {
	testString := "test"
	tests := []struct {
		name    string
		config  ElasticsearchConfig
		isEmpty bool
	}{
		{
			name:    "zero valued config, is empty",
			config:  ElasticsearchConfig{},
			isEmpty: true,
		},
		{
			name: "config with empty plugins, is empty",
			config: ElasticsearchConfig{
				Plugins: []string{},
			},
			isEmpty: true,
		},
		{
			name: "config with non-empty plugins, is non-empty",
			config: ElasticsearchConfig{
				Plugins: []string{"s3"},
			},
			isEmpty: false,
		},
		{
			name: "config with non-empty image, is non-empty",
			config: ElasticsearchConfig{
				DockerImage: &testString,
			},
			isEmpty: false,
		},
		{
			name: "config with non-empty user settings json, is non-empty",
			config: ElasticsearchConfig{
				UserSettingsJson: &testString,
			},
			isEmpty: false,
		},
		{
			name: "config with non-empty user settings override json, is non-empty",
			config: ElasticsearchConfig{
				UserSettingsOverrideJson: &testString,
			},
			isEmpty: false,
		},
		{
			name: "config with non-empty user settings yaml, is non-empty",
			config: ElasticsearchConfig{
				UserSettingsYaml: &testString,
			},
			isEmpty: false,
		},
		{
			name: "config with non-empty user settings override yaml, is non-empty",
			config: ElasticsearchConfig{
				UserSettingsOverrideYaml: &testString,
			},
			isEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.isEmpty, tt.config.IsEmpty())
		})
	}
}
