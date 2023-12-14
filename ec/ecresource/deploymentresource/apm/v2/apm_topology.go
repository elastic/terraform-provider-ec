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
	"context"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/topology/v1"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	minimumApmSize = 512
)

func readApmTopology(in *models.ApmTopologyElement) (*v1.Topology, error) {
	var top v1.Topology

	if in.InstanceConfigurationID != "" {
		top.InstanceConfigurationId = &in.InstanceConfigurationID
	}

	top.InstanceConfigurationVersion = ec.Int(int(in.InstanceConfigurationVersion))

	if in.Size != nil {
		top.Size = ec.String(util.MemoryToState(*in.Size.Value))
		top.SizeResource = ec.String(*in.Size.Resource)
	}

	top.ZoneCount = int(in.ZoneCount)

	return &top, nil
}

func readApmTopologies(in []*models.ApmTopologyElement) (v1.Topologies, error) {
	topologies := make([]v1.Topology, 0, len(in))

	for _, model := range in {
		if model.Size == nil || model.Size.Value == nil || *model.Size.Value == 0 {
			continue
		}

		topology, err := readApmTopology(model)
		if err != nil {
			return nil, nil
		}

		topologies = append(topologies, *topology)
	}

	return topologies, nil
}

// defaultApmTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is smaller than the
// deployment template default, the same is done on the ZoneCount.
func defaultApmTopology(topology []*models.ApmTopologyElement) []*models.ApmTopologyElement {
	for _, t := range topology {
		if *t.Size.Value < minimumApmSize {
			t.Size.Value = ec.Int32(minimumApmSize)
		}
		if t.ZoneCount < utils.MinimumZoneCount {
			t.ZoneCount = utils.MinimumZoneCount
		}
	}

	return topology
}

func apmTopologyPayload(ctx context.Context, topology v1.TopologyTF, model *models.ApmTopologyElement) (*models.ApmTopologyElement, diag.Diagnostics) {

	if topology.InstanceConfigurationId.ValueString() != "" {
		model.InstanceConfigurationID = topology.InstanceConfigurationId.ValueString()
	}

	if !(topology.InstanceConfigurationVersion.IsUnknown() || topology.InstanceConfigurationVersion.IsNull()) {
		model.InstanceConfigurationVersion = int32(topology.InstanceConfigurationVersion.ValueInt64())
	}

	size, err := converters.ParseTopologySizeTypes(topology.Size, topology.SizeResource)

	var diags diag.Diagnostics
	if err != nil {
		diags.AddError("size parsing error", err.Error())
		return nil, diags
	}

	if size != nil {
		model.Size = size
	}

	if topology.ZoneCount.ValueInt64() > 0 {
		model.ZoneCount = int32(topology.ZoneCount.ValueInt64())
	}

	return model, nil
}
