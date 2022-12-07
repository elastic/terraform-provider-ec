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
	"fmt"

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

func integrationsServerTopologyPayload(ctx context.Context, topology topologyv1.TopologyTF, planModels []*models.IntegrationsServerTopologyElement, index int) (*models.IntegrationsServerTopologyElement, diag.Diagnostics) {

	icID := topology.InstanceConfigurationId.Value

	// When a topology element is set but no instance_configuration_id
	// is set, then obtain the instance_configuration_id from the topology
	// element.
	if icID == "" && index < len(planModels) {
		icID = planModels[index].InstanceConfigurationID
	}

	var diags diag.Diagnostics

	size, err := converters.ParseTopologySizeTF(topology.Size, topology.SizeResource)
	if err != nil {
		diags.AddError("parse topology error", err.Error())
		return nil, diags
	}

	elem, err := matchIntegrationsServerTopology(icID, planModels)
	if err != nil {
		diags.AddError("integrations_server topology payload error", err.Error())
		return nil, diags
	}

	if size != nil {
		elem.Size = size
	}

	if topology.ZoneCount.Value > 0 {
		elem.ZoneCount = int32(topology.ZoneCount.Value)
	}

	return elem, nil
}

func matchIntegrationsServerTopology(id string, topologies []*models.IntegrationsServerTopologyElement) (*models.IntegrationsServerTopologyElement, error) {
	for _, t := range topologies {
		if t.InstanceConfigurationID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf(
		`invalid instance_configuration_id: "%s" doesn't match any of the deployment template instance configurations`,
		id,
	)
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

	if in.Size != nil {
		top.Size = ec.String(util.MemoryToState(*in.Size.Value))
		top.SizeResource = ec.String(*in.Size.Resource)
	}

	top.ZoneCount = int(in.ZoneCount)

	return &top, nil
}
