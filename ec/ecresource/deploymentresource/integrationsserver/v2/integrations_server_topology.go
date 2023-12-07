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
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	topologyv1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/topology/v1"
)

const (
	minimumIntegrationsServerSize = 1024
)

func integrationsServerTopologyPayload(ctx context.Context, topology topologyv1.TopologyTF, model *models.IntegrationsServerTopologyElement) (*models.IntegrationsServerTopologyElement, diag.Diagnostics) {

	if topology.InstanceConfigurationId.ValueString() != "" {
		model.InstanceConfigurationID = topology.InstanceConfigurationId.ValueString()
	}

	if topology.InstanceConfigurationVersion.ValueInt64() > 0 {
		model.InstanceConfigurationVersion = int32(topology.InstanceConfigurationVersion.ValueInt64())
	}

	var diags diag.Diagnostics

	size, err := converters.ParseTopologySizeTypes(topology.Size, topology.SizeResource)
	if err != nil {
		diags.AddError("parse topology error", err.Error())
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

// DefaultIntegrationsServerTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is smaller than the
// deployment template default, the same is done on the ZoneCount.
func defaultIntegrationsServerTopology(topology []*models.IntegrationsServerTopologyElement) []*models.IntegrationsServerTopologyElement {
	for _, t := range topology {
		if *t.Size.Value < minimumIntegrationsServerSize {
			t.Size.Value = ec.Int32(minimumIntegrationsServerSize)
		}
		if t.ZoneCount < utils.MinimumZoneCount {
			t.ZoneCount = utils.MinimumZoneCount
		}
	}

	return topology
}

func readIntegrationsServerTopologies(in []*models.IntegrationsServerTopologyElement) (topologyv1.Topologies, error) {
	if len(in) == 0 {
		return nil, nil
	}

	tops := make(topologyv1.Topologies, 0, len(in))
	for _, model := range in {
		if model.Size == nil || model.Size.Value == nil || *model.Size.Value == 0 {
			continue
		}

		top, err := readIntegrationsServerTopology(model)
		if err != nil {
			return nil, err
		}

		tops = append(tops, *top)
	}

	return tops, nil
}

func readIntegrationsServerTopology(in *models.IntegrationsServerTopologyElement) (*topologyv1.Topology, error) {
	var top topologyv1.Topology

	if in.InstanceConfigurationID != "" {
		top.InstanceConfigurationId = &in.InstanceConfigurationID
	}

	top.InstanceConfigurationVersion = int(in.InstanceConfigurationVersion)

	if in.Size != nil {
		top.Size = ec.String(util.MemoryToState(*in.Size.Value))
		top.SizeResource = ec.String(*in.Size.Resource)
	}

	top.ZoneCount = int(in.ZoneCount)

	return &top, nil
}
