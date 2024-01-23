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

package acc

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Test ML tier autoscaling.
// Machine Learning nodes can be auto-scaled exclusively, while the data tier remains being managed manually.
//
// This feature leverages `autoscaling_tier_override` parameter within the ML topology element of the API payload.
func TestAccDeploymentWithMLOnlyAutoscale(t *testing.T) {

	resourceName := "ec_deployment.autoscale_ml"
	initialTfConfigWithMlAutoscale := "testdata/deployment_autoscale_ml.tf"
	tfConfigWithMlAutoscaleDisabled := "testdata/deployment_autoscale_ml_2.tf"
	testID := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomName := prefix + testID

	cfgF := func(cfg string) string {
		return fixtureAccDeploymentResourceBasic(
			t, cfg, randomName, getRegion(), defaultTemplate,
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfgF(initialTfConfigWithMlAutoscale),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.autoscale", "false"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.ml.size", "0g"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.ml.size_resource", "memory"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.ml.zone_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.ml.autoscaling.min_size", "0g"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.ml.autoscaling.autoscale", "true"),
				),
			},
			{
				Config: cfgF(tfConfigWithMlAutoscaleDisabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.autoscale", "false"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.ml.size", "0g"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.ml.size_resource", "memory"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.ml.zone_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.ml.autoscaling.min_size", "0g"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.ml.autoscaling.autoscale", "false"),
				),
			},
		},
	})
}
