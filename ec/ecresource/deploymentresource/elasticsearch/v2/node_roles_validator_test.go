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

package v2_test

import (
	"context"
	"testing"

	deploymentv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestNodeRolesValidator_ValidateSet(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		attrValue types.Set
		isValid   bool
	}{
		{
			name:      "should treat null attribute values as valid",
			version:   utils.DataTiersVersion.String(),
			attrValue: types.SetNull(types.StringType),
			isValid:   true,
		},
		{
			name:      "should treat unknown attribute values as valid",
			version:   utils.DataTiersVersion.String(),
			attrValue: types.SetUnknown(types.StringType),
			isValid:   true,
		},
		{
			name:      "should fail if the deployment version is lt the threshold and the attribute is set",
			version:   "7.0.0",
			attrValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("hot")}),
			isValid:   false,
		},
		{
			name:      "should pass if the deployment version is gte the threshold and the attribute is set",
			version:   utils.DataTiersVersion.String(),
			attrValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("hot")}),
			isValid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := v2.VersionSupportsNodeRoles()
			config := util.TfTypesValueFromGoTypeValue(t, &deploymentv2.Deployment{
				Version: tt.version,
			}, deploymentv2.DeploymentSchema().Type())
			resp := validator.SetResponse{}
			v.ValidateSet(context.Background(), validator.SetRequest{
				ConfigValue: tt.attrValue,
				Config: tfsdk.Config{
					Raw:    config,
					Schema: deploymentv2.DeploymentSchema(),
				},
			}, &resp)

			require.Equal(t, tt.isValid, !resp.Diagnostics.HasError())
		})
	}
}
