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
	"errors"
	"fmt"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/apmstate"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/appsearchstate"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearchstate"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearchstate"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/kibanastate"
)

func modelToState(d *schema.ResourceData, res *models.DeploymentGetResponse) error {
	if err := d.Set("name", res.Name); err != nil {
		return err
	}

	if res.Resources != nil {
		dt, err := getDeploymentTemplateID(res.Resources)
		if err != nil {
			return err
		}

		if err := d.Set("deployment_template_id", dt); err != nil {
			return err
		}

		esFlattened := elasticsearchstate.FlattenResources(res.Resources.Elasticsearch, *res.Name)
		if err := d.Set("elasticsearch", esFlattened); err != nil {
			return err
		}

		kibanaFlattened := kibanastate.FlattenResources(res.Resources.Kibana, *res.Name)
		if err := d.Set("kibana", kibanaFlattened); err != nil {
			return err
		}

		apmFlattened := apmstate.FlattenResources(res.Resources.Apm, *res.Name)
		if err := d.Set("apm", apmFlattened); err != nil {
			return err
		}

		appsearchFlattened := appsearchstate.FlattenResources(res.Resources.Appsearch, *res.Name)
		if err := d.Set("appsearch", appsearchFlattened); err != nil {
			return err
		}

		enterpriseSearchFlattened := enterprisesearchstate.FlattenResources(res.Resources.EnterpriseSearch, *res.Name)
		if err := d.Set("enterprise_search", enterpriseSearchFlattened); err != nil {
			return err
		}
	}

	return nil
}

func getDeploymentTemplateID(res *models.DeploymentResources) (string, error) {
	var deploymentTemplateID string
	var foundTemplates []string
	for _, esRes := range res.Elasticsearch {
		if elasticsearchstate.IsCurrentPlanEmpty(esRes) {
			continue
		}

		var emptyDT = esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate == nil
		if emptyDT {
			continue
		}

		if deploymentTemplateID == "" {
			deploymentTemplateID = *esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate.ID
		}

		foundTemplates = append(foundTemplates,
			*esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate.ID,
		)
	}

	if deploymentTemplateID == "" {
		return "", errors.New("failed to obtain the deployment template id")
	}

	if len(foundTemplates) > 1 {
		return "", fmt.Errorf(
			"there are more than 1 deployment templates specified on the deployment: \"%s\"", strings.Join(foundTemplates, ", "),
		)
	}

	return deploymentTemplateID, nil
}

// parseCredentials parses the Create or Update response Resources populating
// credential settings in the Terraform state if the keys are found, currently
// populates the following credentials in plain text:
// * Elasticsearch username and Password
func parseCredentials(d *schema.ResourceData, resources []*models.DeploymentResource) error {
	var merr = multierror.NewPrefixed("failed parsing credentials")
	for _, res := range resources {
		// Parse ES credentials
		if creds := res.Credentials; creds != nil {
			if creds.Username != nil && *creds.Username != "" {
				if err := d.Set("elasticsearch_username", *creds.Username); err != nil {
					merr = merr.Append(err)
				}
			}

			if creds.Password != nil && *creds.Password != "" {
				if err := d.Set("elasticsearch_password", *creds.Password); err != nil {
					merr = merr.Append(err)
				}
			}
		}
	}

	return merr.ErrorOrNil()
}
