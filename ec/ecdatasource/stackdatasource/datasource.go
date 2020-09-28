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

package stackdatasource

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/stackapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DataSource returns the ec_deployment data source schema.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: read,

		Schema: newSchema(),

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.API)
	region := d.Get("region").(string)

	res, err := stackapi.List(stackapi.ListParams{
		API:    client,
		Region: region,
	})
	if err != nil {
		return diag.FromErr(
			multierror.NewPrefixed("failed retrieving the specified stack version", err),
		)
	}

	versionExpr := d.Get("version_regex").(string)
	version := d.Get("version").(string)
	lock := d.Get("lock").(bool)
	stack, err := stackFromFilters(versionExpr, version, lock, res.Stacks)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Id() == "" {
		d.SetId(strconv.Itoa(schema.HashString(version)))
	}

	if err := modelToState(d, stack); err != nil {
		diag.FromErr(err)
	}

	return nil
}

func stackFromFilters(expr, version string, locked bool, stacks []*models.StackVersionConfig) (*models.StackVersionConfig, error) {
	if expr == "latest" && locked && version != "" {
		expr = version
	}

	if expr == "latest" {
		return stacks[0], nil
	}

	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile the version_regex: %w", err)
	}

	for _, stack := range stacks {
		if re.MatchString(stack.Version) {
			return stack, nil
		}
	}

	return nil, fmt.Errorf(`failed to obtain a stack version matching "%s": `+
		`please specify a valid version_regex`, expr,
	)
}

func modelToState(d *schema.ResourceData, stack *models.StackVersionConfig) error {
	if stack == nil {
		return nil
	}

	if err := d.Set("version", stack.Version); err != nil {
		return err
	}

	if stack.Accessible != nil {
		if err := d.Set("accessible", *stack.Accessible); err != nil {
			return err
		}
	}

	if err := d.Set("min_upgradable_from", stack.MinUpgradableFrom); err != nil {
		return err
	}

	if len(stack.UpgradableTo) > 0 {
		if err := d.Set("upgradable_to", stack.UpgradableTo); err != nil {
			return err
		}
	}

	if stack.Whitelisted != nil {
		if err := d.Set("allowlisted", *stack.Whitelisted); err != nil {
			return err
		}
	}

	if err := d.Set("apm", flattenApmResources(stack.Apm)); err != nil {
		return err
	}

	if err := d.Set("elasticsearch", flattenElasticsearchResources(stack.Elasticsearch)); err != nil {
		return err
	}

	if err := d.Set("enterprise_search", flattenEnterpriseSearchResources(stack.EnterpriseSearch)); err != nil {
		return err
	}

	if err := d.Set("kibana", flattenKibanaResources(stack.Kibana)); err != nil {
		return err
	}

	return nil
}
