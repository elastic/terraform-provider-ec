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
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/enterprisesearch/v1"
	"github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	minimumApmSize              = 512
	minimumEnterpriseSearchSize = 2048
)

type EnterpriseSearchTopologies v1.EnterpriseSearchTopologies

func ReadEnterpriseSearchTopology(in *models.EnterpriseSearchTopologyElement) (*v1.EnterpriseSearchTopology, error) {
	var topology v1.EnterpriseSearchTopology

	topology.InstanceConfigurationId = ec.String(in.InstanceConfigurationID)

	if in.Size != nil {
		topology.Size = ec.String(util.MemoryToState(*in.Size.Value))
		topology.SizeResource = in.Size.Resource
	}

	if nt := in.NodeType; nt != nil {
		if nt.Appserver != nil {
			topology.NodeTypeAppserver = nt.Appserver
		}

		if nt.Connector != nil {
			topology.NodeTypeConnector = nt.Connector
		}

		if nt.Worker != nil {
			topology.NodeTypeWorker = nt.Worker
		}
	}

	topology.ZoneCount = int(in.ZoneCount)

	return &topology, nil
}

func ReadEnterpriseSearchTopologies(in []*models.EnterpriseSearchTopologyElement) (EnterpriseSearchTopologies, error) {
	if len(in) == 0 {
		return nil, nil
	}

	topologies := make(EnterpriseSearchTopologies, 0, len(in))
	for _, model := range in {
		if model.Size == nil || model.Size.Value == nil || *model.Size.Value == 0 {
			continue
		}

		topology, err := ReadEnterpriseSearchTopology(model)
		if err != nil {
			return nil, err
		}

		topologies = append(topologies, *topology)
	}

	return topologies, nil
}

func enterpriseSearchTopologyPayload(ctx context.Context, topology v1.EnterpriseSearchTopologyTF, planModels []*models.EnterpriseSearchTopologyElement, index int) (*models.EnterpriseSearchTopologyElement, diag.Diagnostics) {
	var diags diag.Diagnostics

	icID := topology.InstanceConfigurationId.Value

	// When a topology element is set but no instance_configuration_id
	// is set, then obtain the instance_configuration_id from the topology
	// element.
	if icID == "" && index < len(planModels) {
		icID = planModels[index].InstanceConfigurationID
	}

	elem, err := matchEssTopology(icID, planModels)
	if err != nil {
		diags.AddError("cannot match enterprise search topology", err.Error())
		return nil, diags
	}

	size, err := converters.ParseTopologySizeTF(topology.Size, topology.SizeResource)

	if err != nil {
		diags.AddError("failed parse enterprise search topology size", err.Error())
		return nil, diags
	}

	// Since Enterprise Search is not enabled by default in the template,
	// if the size == nil, it means that the size hasn't been specified in
	// the definition.
	if size == nil {
		size = &models.TopologySize{
			Resource: ec.String("memory"),
			Value:    ec.Int32(minimumEnterpriseSearchSize),
		}
	}

	elem.Size = size

	if topology.ZoneCount.Value > 0 {
		elem.ZoneCount = int32(topology.ZoneCount.Value)
	}

	return elem, nil
}

// defaultApmTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is smaller than the
// deployment template default, the same is done on the ZoneCount.
func defaultEssTopology(topology []*models.EnterpriseSearchTopologyElement) []*models.EnterpriseSearchTopologyElement {
	for _, t := range topology {
		if *t.Size.Value < minimumEnterpriseSearchSize || *t.Size.Value == 0 {
			t.Size.Value = ec.Int32(minimumEnterpriseSearchSize)
		}
		if t.ZoneCount < utils.MinimumZoneCount {
			t.ZoneCount = utils.MinimumZoneCount
		}
	}

	return topology
}

func matchEssTopology(id string, topologies []*models.EnterpriseSearchTopologyElement) (*models.EnterpriseSearchTopologyElement, error) {
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
