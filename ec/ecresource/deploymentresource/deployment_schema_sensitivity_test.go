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

package deploymentresource_test

import (
	"testing"

	deploymentv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v1"
	deploymentv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"
)

func TestDeploymentUserSettingsAreSensitive(t *testing.T) {
	tests := []struct {
		name          string
		schema        schema.Schema
		expectedCount int
	}{
		{
			name:          "v1",
			schema:        deploymentv1.DeploymentSchema(),
			expectedCount: 24,
		},
		{
			name:          "v2",
			schema:        deploymentv2.DeploymentSchema(),
			expectedCount: 20,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expectedCount, assertSensitiveUserSettings(t, test.schema.Attributes))
		})
	}
}

func assertSensitiveUserSettings(t *testing.T, attributes map[string]schema.Attribute) int {
	t.Helper()

	count := 0
	for name, attribute := range attributes {
		switch name {
		case "user_settings_json", "user_settings_yaml", "user_settings_override_json", "user_settings_override_yaml":
			stringAttribute, ok := attribute.(schema.StringAttribute)
			require.True(t, ok, "%s must remain a string attribute", name)
			require.True(t, stringAttribute.Sensitive, "%s must be sensitive", name)
			count++
		}

		switch nested := attribute.(type) {
		case schema.SingleNestedAttribute:
			count += assertSensitiveUserSettings(t, nested.Attributes)
		case schema.ListNestedAttribute:
			count += assertSensitiveUserSettings(t, nested.NestedObject.Attributes)
		}
	}

	return count
}
