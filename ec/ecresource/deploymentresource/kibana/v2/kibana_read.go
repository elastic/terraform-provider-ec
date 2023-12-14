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

package v2

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

type Kibana struct {
	ElasticsearchClusterRefId    *string       `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                        *string       `tfsdk:"ref_id"`
	ResourceId                   *string       `tfsdk:"resource_id"`
	Region                       *string       `tfsdk:"region"`
	HttpEndpoint                 *string       `tfsdk:"http_endpoint"`
	HttpsEndpoint                *string       `tfsdk:"https_endpoint"`
	InstanceConfigurationId      *string       `tfsdk:"instance_configuration_id"`
	InstanceConfigurationVersion *int          `tfsdk:"instance_configuration_version"`
	Size                         *string       `tfsdk:"size"`
	SizeResource                 *string       `tfsdk:"size_resource"`
	ZoneCount                    int           `tfsdk:"zone_count"`
	Config                       *KibanaConfig `tfsdk:"config"`
}

func ReadKibanas(in []*models.KibanaResourceInfo) (*Kibana, error) {
	for _, model := range in {
		if util.IsCurrentKibanaPlanEmpty(model) || IsKibanaStopped(model) {
			continue
		}

		kibana, err := readKibana(model)
		if err != nil {
			return nil, err
		}

		return kibana, nil
	}

	return nil, nil
}

func readKibana(in *models.KibanaResourceInfo) (*Kibana, error) {
	var kibana Kibana

	kibana.RefId = in.RefID

	kibana.ResourceId = in.Info.ClusterID

	kibana.Region = in.Region

	plan := in.Info.PlanInfo.Current.Plan
	var err error

	topologies, err := readKibanaTopologies(plan.ClusterTopology)
	if err != nil {
		return nil, err
	}

	if len(topologies) > 0 {
		kibana.InstanceConfigurationId = topologies[0].InstanceConfigurationId
		kibana.InstanceConfigurationVersion = topologies[0].InstanceConfigurationVersion
		kibana.Size = topologies[0].Size
		kibana.SizeResource = topologies[0].SizeResource
		kibana.ZoneCount = topologies[0].ZoneCount
	}

	kibana.ElasticsearchClusterRefId = in.ElasticsearchClusterRefID

	kibana.HttpEndpoint, kibana.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)

	config, err := readKibanaConfig(plan.Kibana)
	if err != nil {
		return nil, err
	}

	kibana.Config = config

	return &kibana, nil
}

// IsKibanaStopped returns true if the resource is stopped.
func IsKibanaStopped(res *models.KibanaResourceInfo) bool {
	return res == nil || res.Info == nil || res.Info.Status == nil ||
		*res.Info.Status == "stopped"
}
