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
	"context"
	"fmt"
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/trafficfilterapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/deployment/v2"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/assert"
)

func Test_handleTrafficFilterChange(t *testing.T) {
	deploymentID := "deployment_unique_id"

	type args struct {
		plan  []string
		state []string
	}

	tests := []struct {
		name       string
		args       args
		getRule    func(trafficfilterapi.GetParams) (*models.TrafficFilterRulesetInfo, error)
		createRule func(params trafficfilterapi.CreateAssociationParams) error
		deleteRule func(params trafficfilterapi.DeleteAssociationParams) error
	}{
		{
			name: "should not call the association API when plan and state contain same rules",
			args: args{
				plan:  []string{"rule1"},
				state: []string{"rule1"},
			},
			getRule: func(trafficfilterapi.GetParams) (*models.TrafficFilterRulesetInfo, error) {
				err := "GetRule function SHOULD NOT be called"
				return nil, fmt.Errorf("%v", err)
			},
			createRule: func(params trafficfilterapi.CreateAssociationParams) error {
				err := "CreateRule function SHOULD NOT be called"
				return fmt.Errorf("%v", err)
			},
			deleteRule: func(params trafficfilterapi.DeleteAssociationParams) error {
				err := "DeleteRule function SHOULD NOT be called"
				return fmt.Errorf("%v", err)
			},
		},

		{
			name: "should add rule when plan contains it and state doesn't contain it",
			args: args{
				plan:  []string{"rule1", "rule2"},
				state: []string{"rule1"},
			},
			getRule: func(trafficfilterapi.GetParams) (*models.TrafficFilterRulesetInfo, error) {
				return &models.TrafficFilterRulesetInfo{}, nil
			},
			createRule: func(params trafficfilterapi.CreateAssociationParams) error {
				assert.Equal(t, "rule2", params.ID)
				return nil
			},
			deleteRule: func(params trafficfilterapi.DeleteAssociationParams) error {
				err := "DeleteRule function SHOULD NOT be called"
				return fmt.Errorf("%v", err)
			},
		},

		{
			name: "should not add rule when plan contains it and state doesn't contain it but the association already exists",
			args: args{
				plan:  []string{"rule1", "rule2"},
				state: []string{"rule1"},
			},
			getRule: func(trafficfilterapi.GetParams) (*models.TrafficFilterRulesetInfo, error) {
				return &models.TrafficFilterRulesetInfo{
					Associations: []*models.FilterAssociation{
						{
							ID:         &deploymentID,
							EntityType: ec.String("deployment"),
						},
					},
				}, nil
			},
			createRule: func(params trafficfilterapi.CreateAssociationParams) error {
				err := "CreateRule function SHOULD NOT be called"
				return fmt.Errorf("%v", err)
			},
			deleteRule: func(params trafficfilterapi.DeleteAssociationParams) error {
				err := "DeleteRule function SHOULD NOT be called"
				return fmt.Errorf("%v", err)
			},
		},

		{
			name: "should delete rule when plan doesn't contain it and state does contain it",
			args: args{
				plan:  []string{"rule1"},
				state: []string{"rule1", "rule2"},
			},
			getRule: func(trafficfilterapi.GetParams) (*models.TrafficFilterRulesetInfo, error) {
				return &models.TrafficFilterRulesetInfo{
					Associations: []*models.FilterAssociation{
						{
							ID:         &deploymentID,
							EntityType: ec.String("deployment"),
						},
					},
				}, nil
			},
			createRule: func(params trafficfilterapi.CreateAssociationParams) error {
				err := "CreateRule function SHOULD NOT be called"
				return fmt.Errorf("%v", err)
			},
			deleteRule: func(params trafficfilterapi.DeleteAssociationParams) error {
				assert.Equal(t, "rule2", params.ID)
				return nil
			},
		},

		{
			name: "should not delete rule when plan doesn't contain it and state does contain it but the association is already gone",
			args: args{
				plan:  []string{"rule1"},
				state: []string{"rule1", "rule2"},
			},
			getRule: func(trafficfilterapi.GetParams) (*models.TrafficFilterRulesetInfo, error) {
				return &models.TrafficFilterRulesetInfo{}, nil
			},
			createRule: func(params trafficfilterapi.CreateAssociationParams) error {
				err := "CreateRule function SHOULD NOT be called"
				return fmt.Errorf("%v", err)
			},
			deleteRule: func(params trafficfilterapi.DeleteAssociationParams) error {
				err := "DeleteRule function SHOULD NOT be called"
				return fmt.Errorf("%v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getRule := deploymentresource.GetAssociation
			createRule := deploymentresource.CreateAssociation
			deleteRule := deploymentresource.DeleteAssociation

			defer func() {
				deploymentresource.GetAssociation = getRule
				deploymentresource.CreateAssociation = createRule
				deploymentresource.DeleteAssociation = deleteRule
			}()

			deploymentresource.GetAssociation = tt.getRule
			deploymentresource.CreateAssociation = tt.createRule
			deploymentresource.DeleteAssociation = tt.deleteRule

			plan := v2.Deployment{
				Id:            deploymentID,
				TrafficFilter: tt.args.plan,
			}

			var planTF v2.DeploymentTF
			diags := tfsdk.ValueFrom(context.Background(), &plan, v2.DeploymentSchema().Type(), &planTF)
			assert.Nil(t, diags)

			filters, diags := deploymentresource.HandleTrafficFilterChange(context.Background(), nil, planTF, tt.args.state)

			assert.Nil(t, diags)
			assert.ElementsMatch(t, tt.args.plan, filters)
		})

	}

}
