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
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDeployment_docker_image_override(t *testing.T) {
	resName := "ec_deployment.docker_image"
	randomName := prefix + "docker_image_" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	cfgF := func(cfg string) string {
		t.Helper()
		requiresAPIConn(t)

		b, err := os.ReadFile(cfg)
		if err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf(string(b),
			randomName, "gcp-us-west2", setDefaultTemplate("gcp-us-west2", defaultTemplate),
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactory,
		CheckDestroy:             testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfgF("testdata/deployment_docker_image_override.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.config.docker_image", "docker.elastic.co/cloud-ci/elasticsearch:7.15.0-SNAPSHOT"),
					resource.TestCheckResourceAttr(resName, "kibana.config.docker_image", "docker.elastic.co/cloud-ci/kibana:7.15.0-SNAPSHOT"),
					resource.TestCheckResourceAttr(resName, "apm.config.docker_image", "docker.elastic.co/cloud-ci/apm:7.15.0-SNAPSHOT"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.config.docker_image", "docker.elastic.co/cloud-ci/enterprise-search:7.15.0-SNAPSHOT"),
				),
			},
		},
	})
}
