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

package deploymentresource

import (
	"context"
	"errors"
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Setting this variable here so that it is parsed at compile time in case
// any errors are thrown, they are at compile time not when the user runs it.
var ilmVersion = semver.MustParse("6.6.0")

// imports a deployment limitting the allowed version to 6.6.0 or higher.
// TODO: It might be desired to provide the ability to import a deployment
// specifying key:value pairs of secrets to populate as part of the
// import with an implementation of schema.StateContextFunc.
func importFunc(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*api.API)
	res, err := deploymentapi.Get(deploymentapi.GetParams{
		API:          client,
		DeploymentID: d.Id(),
		QueryParams: deputil.QueryParams{
			ShowPlans: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(res.Resources.Elasticsearch) == 0 {
		return nil, errors.New(
			"invalid deployment: deployment has no elasticsearch resources",
		)
	}

	v, err := semver.New(
		res.Resources.Elasticsearch[0].Info.PlanInfo.Current.Plan.Elasticsearch.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to parse deployment version: %w", err)
	}

	if v.LT(ilmVersion) {
		return nil, fmt.Errorf(
			`invalid deployment version "%s": minimum supported version is "%s"`,
			v.String(), ilmVersion.String(),
		)
	}

	return []*schema.ResourceData{d}, nil
}
