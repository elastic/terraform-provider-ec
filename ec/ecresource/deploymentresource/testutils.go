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
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type resDataParams struct {
	Resources map[string]interface{}
	ID        string
}

func newResourceData(t *testing.T, params resDataParams) *schema.ResourceData {
	raw := schema.TestResourceDataRaw(t, NewSchema(), params.Resources)
	raw.SetId(params.ID)

	return raw
}

func newSampleDeployment() map[string]interface{} {
	return map[string]interface{}{
		"name":                   "my_deployment_name",
		"deployment_template_id": "aws-io-optimized",
		"elasticsearch":          []interface{}{newElasticsearchSample()},
		"kibana":                 []interface{}{newKibanaSample()},
		"apm":                    []interface{}{newApmSample()},
		"appsearch":              []interface{}{newAppsearchSample()},
		"enterprise_search":      []interface{}{newEnterpriseSearchSample()},
	}
}

func newElasticsearchSample() map[string]interface{} {
	return map[string]interface{}{
		"display_name": "some-name",
		"ref_id":       "main-elasticsearch",
		"resource_id":  mock.ValidClusterID,
		"version":      "7.7.0",
		"region":       "some-region",
		"topology": []interface{}{
			map[string]interface{}{
				"instance_configuration_id": "aws.data.highio.i3",
				"memory_per_node":           "2g",
				"node_type_data":            true,
				"node_type_ingest":          true,
				"node_type_master":          true,
				"node_type_ml":              false,
				"zone_count":                1,
			},
		},
		"monitoring_settings": []interface{}{
			map[string]interface{}{"target_cluster_id": "some"},
		},
	}
}

func newKibanaSample() map[string]interface{} {
	return map[string]interface{}{
		"elasticsearch_cluster_ref_id": "main-elasticsearch",
		"display_name":                 "some-kibana-name",
		"ref_id":                       "main-kibana",
		"resource_id":                  mock.ValidClusterID,
		"version":                      "7.7.0",
		"region":                       "some-region",
		"topology": []interface{}{
			map[string]interface{}{
				"instance_configuration_id": "aws.kibana.r4",
				"memory_per_node":           "1g",
				"zone_count":                1,
			},
		},
	}
}

func newApmSample() map[string]interface{} {
	return map[string]interface{}{
		"elasticsearch_cluster_ref_id": "main-elasticsearch",
		"display_name":                 "some-apm-name",
		"ref_id":                       "main-apm",
		"resource_id":                  mock.ValidClusterID,
		"version":                      "7.7.0",
		"region":                       "some-region",
		// Reproduces the case where the default fields are set.
		"config": []interface{}{map[string]interface{}{
			"debug_enabled": false,
		}},
		"topology": []interface{}{map[string]interface{}{
			"instance_configuration_id": "aws.apm.r4",
			"memory_per_node":           "0.5g",
			"zone_count":                1,
			"config": []interface{}{map[string]interface{}{
				"debug_enabled": false,
			}},
		}},
	}
}

func newAppsearchSample() map[string]interface{} {
	return map[string]interface{}{
		"elasticsearch_cluster_ref_id": "main-elasticsearch",
		"display_name":                 "some-appsearch-name",
		"ref_id":                       "main-appsearch",
		"resource_id":                  mock.ValidClusterID,
		"version":                      "7.7.0",
		"region":                       "some-region",
		"topology": []interface{}{
			map[string]interface{}{
				"instance_configuration_id": "aws.appsearch.m5",
				"memory_per_node":           "2g",
				"zone_count":                1,
				"node_type_appserver":       true,
				"node_type_worker":          true,
			},
		},
	}
}

func newEnterpriseSearchSample() map[string]interface{} {
	return map[string]interface{}{
		"elasticsearch_cluster_ref_id": "main-elasticsearch",
		"display_name":                 "some-enterprise_search-name",
		"ref_id":                       "main-enterprise_search",
		"resource_id":                  mock.ValidClusterID,
		"version":                      "7.7.0",
		"region":                       "some-region",
		"topology": []interface{}{
			map[string]interface{}{
				"instance_configuration_id": "aws.enterprisesearch.m5",
				"memory_per_node":           "2g",
				"zone_count":                1,
				"node_type_appserver":       true,
				"node_type_connector":       true,
				"node_type_worker":          true,
			},
		},
	}
}
