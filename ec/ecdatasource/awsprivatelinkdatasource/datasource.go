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

package awsprivatelinkdatasource

import (
	"context"
	_ "embed"
	"encoding/json"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//go:embed regionPrivateLinkMap.json
var privateLinkDataJson string

// DataSource returns the ec_aws_privatelink_endpoint data source schema.
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
	var privateLinkData map[string]struct {
		VpcServiceName string   `json:"vpc_service_name"`
		DomainName     string   `json:"domain_name"`
		ZoneIds        []string `json:"zone_ids"`
	}

	if err := json.Unmarshal([]byte(privateLinkDataJson), &privateLinkData); err != nil {
		return diag.FromErr(err)
	}

	region, ok := d.Get("region").(string)
	if !ok {
		return diag.Errorf("a region is required to lookup a privatelink endpoint")
	}

	if d.Id() == "" {
		d.SetId(strconv.Itoa(schema.HashString(region)))
	}

	regionLink, ok := privateLinkData[region]
	if !ok {
		return diag.Errorf("could not find a privatelink endpoint for region: %s", region)
	}

	if err := d.Set("vpc_service_name", regionLink.VpcServiceName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("domain_name", regionLink.DomainName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("zone_ids", regionLink.ZoneIds); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
