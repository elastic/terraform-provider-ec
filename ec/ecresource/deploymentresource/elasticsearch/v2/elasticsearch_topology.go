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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deploymentsize"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource/deploymentresource/elasticsearch/v1"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopologyTF struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
	NodeTypeData            types.String `tfsdk:"node_type_data"`
	NodeTypeMaster          types.String `tfsdk:"node_type_master"`
	NodeTypeIngest          types.String `tfsdk:"node_type_ingest"`
	NodeTypeMl              types.String `tfsdk:"node_type_ml"`
	NodeRoles               types.Set    `tfsdk:"node_roles"`
	Autoscaling             types.Object `tfsdk:"autoscaling"`
}

type ElasticsearchTopology struct {
	id                      string
	InstanceConfigurationId *string                           `tfsdk:"instance_configuration_id"`
	Size                    *string                           `tfsdk:"size"`
	SizeResource            *string                           `tfsdk:"size_resource"`
	ZoneCount               int                               `tfsdk:"zone_count"`
	NodeTypeData            *string                           `tfsdk:"node_type_data"`
	NodeTypeMaster          *string                           `tfsdk:"node_type_master"`
	NodeTypeIngest          *string                           `tfsdk:"node_type_ingest"`
	NodeTypeMl              *string                           `tfsdk:"node_type_ml"`
	NodeRoles               []string                          `tfsdk:"node_roles"`
	Autoscaling             *ElasticsearchTopologyAutoscaling `tfsdk:"autoscaling"`
}

func CreateTierForTest(tierId string, tier ElasticsearchTopology) *ElasticsearchTopology {
	res := tier
	res.id = tierId
	return &res
}

type ElasticsearchTopologyAutoscaling v1.ElasticsearchTopologyAutoscaling

func (topology ElasticsearchTopologyTF) Payload(ctx context.Context, topologyID string, planTopologies []*models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	var diags diag.Diagnostics

	topologyElem, err := matchEsTopologyID(topologyID, planTopologies)
	if err != nil {
		diags.AddError("topology matching error", err.Error())
		return diags
	}

	size, err := converters.ParseTopologySizeTF(topology.Size, topology.SizeResource)
	if err != nil {
		diags.AddError("size parsing error", err.Error())
	}

	if size != nil {
		topologyElem.Size = size
	}

	if topology.ZoneCount.Value > 0 {
		topologyElem.ZoneCount = int32(topology.ZoneCount.Value)
	}

	if err := topology.ParseLegacyNodeType(topologyElem.NodeType); err != nil {
		diags.AddError("topology legacy node type error", err.Error())
	}

	var nodeRoles []string
	ds := topology.NodeRoles.ElementsAs(ctx, &nodeRoles, true)
	diags.Append(ds...)

	if !ds.HasError() && len(nodeRoles) > 0 {
		topologyElem.NodeRoles = nodeRoles
		topologyElem.NodeType = nil
	}

	diags.Append(ElasticsearchTopologyAutoscalingPayload(ctx, topology.Autoscaling, topologyID, topologyElem)...)

	diags = append(diags, ds...)

	return diags
}

func ReadElasticsearchTopologies(in *models.ElasticsearchClusterPlan) (ElasticsearchTopologies, error) {
	if len(in.ClusterTopology) == 0 {
		return nil, nil
	}

	tops := make([]ElasticsearchTopology, 0, len(in.ClusterTopology))

	for _, model := range in.ClusterTopology {
		// if !v1.IsPotentiallySizedTopology(model, in.AutoscalingEnabled != nil && *in.AutoscalingEnabled) {
		// 	continue
		// }

		topology, err := ReadElasticsearchTopology(model)
		if err != nil {
			return nil, err
		}
		tops = append(tops, *topology)
	}

	return tops, nil
}

func ReadElasticsearchTopology(model *models.ElasticsearchClusterTopologyElement) (*ElasticsearchTopology, error) {
	var topology ElasticsearchTopology

	topology.id = model.ID

	if model.InstanceConfigurationID != "" {
		topology.InstanceConfigurationId = &model.InstanceConfigurationID
	}

	if model.Size != nil {
		topology.Size = ec.String(util.MemoryToState(*model.Size.Value))
		topology.SizeResource = model.Size.Resource
	}

	topology.ZoneCount = int(model.ZoneCount)

	if nt := model.NodeType; nt != nil {
		if nt.Data != nil {
			topology.NodeTypeData = ec.String(strconv.FormatBool(*nt.Data))
		}

		if nt.Ingest != nil {
			topology.NodeTypeIngest = ec.String(strconv.FormatBool(*nt.Ingest))
		}

		if nt.Master != nil {
			topology.NodeTypeMaster = ec.String(strconv.FormatBool(*nt.Master))
		}

		if nt.Ml != nil {
			topology.NodeTypeMl = ec.String(strconv.FormatBool(*nt.Ml))
		}
	}

	topology.NodeRoles = model.NodeRoles

	autoscaling, err := ReadElasticsearchTopologyAutoscaling(model)
	if err != nil {
		return nil, err
	}
	topology.Autoscaling = autoscaling

	return &topology, nil
}

func ReadElasticsearchTopologyAutoscaling(topology *models.ElasticsearchClusterTopologyElement) (*ElasticsearchTopologyAutoscaling, error) {
	var a ElasticsearchTopologyAutoscaling

	if ascale := topology.AutoscalingMax; ascale != nil {
		a.MaxSizeResource = ascale.Resource
		a.MaxSize = ec.String(util.MemoryToState(*ascale.Value))
	}

	if ascale := topology.AutoscalingMin; ascale != nil {
		a.MinSizeResource = ascale.Resource
		a.MinSize = ec.String(util.MemoryToState(*ascale.Value))
	}

	if topology.AutoscalingPolicyOverrideJSON != nil {
		b, err := json.Marshal(topology.AutoscalingPolicyOverrideJSON)
		if err != nil {
			return nil, fmt.Errorf("elasticsearch topology %s: unable to persist policy_override_json - %w", topology.ID, err)
		}
		a.PolicyOverrideJson = ec.String(string(b))
	}

	return &a, nil
}

func (topology *ElasticsearchTopologyTF) ParseLegacyNodeType(nodeType *models.ElasticsearchNodeType) error {
	if nodeType == nil {
		return nil
	}

	if topology.NodeTypeData.Value != "" {
		nt, err := strconv.ParseBool(topology.NodeTypeData.Value)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_data value: %w", err)
		}
		nodeType.Data = &nt
	}

	if topology.NodeTypeMaster.Value != "" {
		nt, err := strconv.ParseBool(topology.NodeTypeMaster.Value)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_master value: %w", err)
		}
		nodeType.Master = &nt
	}

	if topology.NodeTypeIngest.Value != "" {
		nt, err := strconv.ParseBool(topology.NodeTypeIngest.Value)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_ingest value: %w", err)
		}
		nodeType.Ingest = &nt
	}

	if topology.NodeTypeMl.Value != "" {
		nt, err := strconv.ParseBool(topology.NodeTypeMl.Value)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_ml value: %w", err)
		}
		nodeType.Ml = &nt
	}

	return nil
}

func (topology *ElasticsearchTopologyTF) HasNodeType() bool {
	return topology.NodeTypeData.Value != "" ||
		topology.NodeTypeIngest.Value != "" ||
		topology.NodeTypeMaster.Value != "" ||
		topology.NodeTypeMl.Value != ""
}

func ObjectToTopology(ctx context.Context, obj types.Object) (*ElasticsearchTopologyTF, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}

	var topology *ElasticsearchTopologyTF

	if diags := tfsdk.ValueAs(ctx, obj, &topology); diags.HasError() {
		return nil, diags
	}

	return topology, nil
}

type ElasticsearchTopologies []ElasticsearchTopology

func (tops ElasticsearchTopologies) Set() map[string]ElasticsearchTopology {
	set := make(map[string]ElasticsearchTopology, len(tops))

	for _, top := range tops {
		set[top.id] = top
	}

	return set
}

func matchEsTopologyID(id string, topologies []*models.ElasticsearchClusterTopologyElement) (*models.ElasticsearchClusterTopologyElement, error) {
	for _, t := range topologies {
		if t.ID == id {
			return t, nil
		}
	}

	topIDs := topologyIDs(topologies)
	for i, id := range topIDs {
		topIDs[i] = "\"" + id + "\""
	}

	return nil, fmt.Errorf(`invalid id ('%s'): valid topology IDs are %s`, id, strings.Join(topIDs, ", "))
}

func topologyIDs(topologies []*models.ElasticsearchClusterTopologyElement) []string {
	var result []string

	for _, topology := range topologies {
		result = append(result, topology.ID)
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func ElasticsearchTopologyAutoscalingPayload(ctx context.Context, autoObj attr.Value, topologyID string, payload *models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	var diag diag.Diagnostics

	if autoObj.IsNull() || autoObj.IsUnknown() {
		return nil
	}

	// it should be only one element if any
	var autoscale v1.ElasticsearchTopologyAutoscalingTF

	if diags := tfsdk.ValueAs(ctx, autoObj, &autoscale); diags.HasError() {
		return diags
	}

	if autoscale == (v1.ElasticsearchTopologyAutoscalingTF{}) {
		return nil
	}

	if !autoscale.MinSize.IsNull() && !autoscale.MinSize.IsUnknown() {
		if payload.AutoscalingMin == nil {
			payload.AutoscalingMin = new(models.TopologySize)
		}

		err := expandAutoscalingDimension(autoscale, payload.AutoscalingMin, autoscale.MinSize, autoscale.MinSizeResource)
		if err != nil {
			diag.AddError("fail to parse autoscale min size", err.Error())
			return diag
		}

		if reflect.DeepEqual(payload.AutoscalingMin, new(models.TopologySize)) {
			payload.AutoscalingMin = nil
		}
	}

	if !autoscale.MaxSize.IsNull() && !autoscale.MaxSize.IsUnknown() {
		if payload.AutoscalingMax == nil {
			payload.AutoscalingMax = new(models.TopologySize)
		}

		err := expandAutoscalingDimension(autoscale, payload.AutoscalingMax, autoscale.MaxSize, autoscale.MaxSizeResource)
		if err != nil {
			diag.AddError("fail to parse autoscale max size", err.Error())
			return diag
		}

		if reflect.DeepEqual(payload.AutoscalingMax, new(models.TopologySize)) {
			payload.AutoscalingMax = nil
		}
	}

	if autoscale.PolicyOverrideJson.Value != "" {
		if err := json.Unmarshal([]byte(autoscale.PolicyOverrideJson.Value),
			&payload.AutoscalingPolicyOverrideJSON,
		); err != nil {
			diag.AddError(fmt.Sprintf("elasticsearch topology %s: unable to load policy_override_json", topologyID), err.Error())
			return diag
		}
	}

	return diag
}

// expandAutoscalingDimension centralises processing of %_size and %_size_resource attributes
// Due to limitations in the Terraform SDK, it's not possible to specify a Default on a Computed schema member
// to work around this limitation, this function will default the %_size_resource attribute to `memory`.
// Without this default, setting autoscaling limits on tiers which do not have those limits in the deployment
// template leads to an API error due to the empty resource field on the TopologySize model.
func expandAutoscalingDimension(autoscale v1.ElasticsearchTopologyAutoscalingTF, model *models.TopologySize, size, sizeResource types.String) error {
	if size.Value != "" {
		val, err := deploymentsize.ParseGb(size.Value)
		if err != nil {
			return err
		}
		model.Value = &val

		if model.Resource == nil {
			model.Resource = ec.String("memory")
		}
	}

	if sizeResource.Value != "" {
		model.Resource = &sizeResource.Value
	}

	return nil
}
