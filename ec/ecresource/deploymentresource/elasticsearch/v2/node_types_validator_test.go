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

	entsearch "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v2"
	obs "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/observability/v2"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"

	apm "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/apm/v2"
	deploymentv2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v2"
	is "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/integrationsserver/v2"
	kibana "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/kibana/v2"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestNodeTypesValidator_ValidateString(t *testing.T) {
	tests := []struct {
		name      string
		version   types.String
		attrValue types.String
		isValid   bool
	}{
		{
			name:      "should treat null attribute values as valid",
			version:   types.StringValue(utils.MinVersionWithoutNodeTypes.String()),
			attrValue: types.StringNull(),
			isValid:   true,
		},
		{
			name:      "should treat unknown attribute values as valid",
			version:   types.StringValue(utils.MinVersionWithoutNodeTypes.String()),
			attrValue: types.StringUnknown(),
			isValid:   true,
		},
		{
			name:      "should fail if the deployment version is gte the threshold and the attribute is set",
			version:   types.StringValue(utils.MinVersionWithoutNodeTypes.String()),
			attrValue: types.StringValue("false"),
			isValid:   false,
		},
		{
			name:      "should pass if the deployment version is lt the threshold and the attribute is set",
			version:   types.StringValue("7.17.9"),
			attrValue: types.StringValue("false"),
			isValid:   true,
		},
		{
			name:      "should treat unknown version as valid",
			version:   types.StringUnknown(),
			attrValue: types.StringValue("true"),
			isValid:   true,
		},
		{
			name:      "should treat null version as valid",
			version:   types.StringNull(),
			attrValue: types.StringValue("true"),
			isValid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := v2.VersionSupportsNodeTypes()
			config := util.TfTypesValueFromGoTypeValue(t,
				&deploymentv2.DeploymentTF{
					Version:            tt.version,
					Elasticsearch:      types.ObjectUnknown(v2.ElasticsearchSchema().GetType().(types.ObjectType).AttrTypes),
					Kibana:             types.ObjectUnknown(kibana.KibanaSchema().GetType().(types.ObjectType).AttrTypes),
					Apm:                types.ObjectUnknown(apm.ApmSchema().GetType().(types.ObjectType).AttrTypes),
					IntegrationsServer: types.ObjectUnknown(is.IntegrationsServerSchema().GetType().(types.ObjectType).AttrTypes),
					EnterpriseSearch:   types.ObjectUnknown(entsearch.EnterpriseSearchSchema().GetType().(types.ObjectType).AttrTypes),
					TrafficFilter:      types.SetUnknown(types.StringType),
					Tags:               types.MapUnknown(types.StringType),
					Observability:      types.ObjectUnknown(obs.ObservabilitySchema().GetType().(types.ObjectType).AttrTypes),
				},
				deploymentv2.DeploymentSchema().Type())

			resp := validator.StringResponse{}
			v.ValidateString(context.Background(), validator.StringRequest{
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
