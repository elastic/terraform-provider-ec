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

type Elasticsearch struct {
	Autoscale        *bool                        `tfsdk:"autoscale"`
	RefId            *string                      `tfsdk:"ref_id"`
	ResourceId       *string                      `tfsdk:"resource_id"`
	Region           *string                      `tfsdk:"region"`
	CloudID          *string                      `tfsdk:"cloud_id"`
	HttpEndpoint     *string                      `tfsdk:"http_endpoint"`
	HttpsEndpoint    *string                      `tfsdk:"https_endpoint"`
	HotTier          *ElasticsearchTopology       `tfsdk:"hot"`
	CoordinatingTier *ElasticsearchTopology       `tfsdk:"coordinating"`
	MasterTier       *ElasticsearchTopology       `tfsdk:"master"`
	WarmTier         *ElasticsearchTopology       `tfsdk:"warm"`
	ColdTier         *ElasticsearchTopology       `tfsdk:"cold"`
	FrozenTier       *ElasticsearchTopology       `tfsdk:"frozen"`
	MlTier           *ElasticsearchTopology       `tfsdk:"ml"`
	Config           *ElasticsearchConfig         `tfsdk:"config"`
	RemoteCluster    ElasticsearchRemoteClusters  `tfsdk:"remote_cluster"`
	SnapshotSource   *ElasticsearchSnapshotSource `tfsdk:"snapshot_source"`
	Extension        ElasticsearchExtensions      `tfsdk:"extension"`
	TrustAccount     ElasticsearchTrustAccounts   `tfsdk:"trust_account"`
	TrustExternal    ElasticsearchTrustExternals  `tfsdk:"trust_external"`
	Strategy         *string                      `tfsdk:"strategy"`
}

func ReadElasticsearches(in []*models.ElasticsearchResourceInfo, remotes *models.RemoteResources) (*Elasticsearch, error) {
	for _, model := range in {
		if util.IsCurrentEsPlanEmpty(model) || IsElasticsearchStopped(model) {
			continue
		}
		es, err := readElasticsearch(model, remotes)
		if err != nil {
			return nil, err
		}
		return es, nil
	}

	return nil, nil
}

func readElasticsearch(in *models.ElasticsearchResourceInfo, remotes *models.RemoteResources) (*Elasticsearch, error) {
	var es Elasticsearch

	if util.IsCurrentEsPlanEmpty(in) || IsElasticsearchStopped(in) {
		return &es, nil
	}

	if in.Info.ClusterID != nil && *in.Info.ClusterID != "" {
		es.ResourceId = in.Info.ClusterID
	}

	if in.RefID != nil && *in.RefID != "" {
		es.RefId = in.RefID
	}

	if in.Region != nil {
		es.Region = in.Region
	}

	plan := in.Info.PlanInfo.Current.Plan
	var err error

	topologies, err := readElasticsearchTopologies(plan)
	if err != nil {
		return nil, err
	}
	es.setTopology(topologies)

	if plan.AutoscalingEnabled != nil {
		es.Autoscale = plan.AutoscalingEnabled
	}

	if meta := in.Info.Metadata; meta != nil && meta.CloudID != "" {
		es.CloudID = &meta.CloudID
	}

	es.HttpEndpoint, es.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)

	es.Config, err = readElasticsearchConfig(plan.Elasticsearch)
	if err != nil {
		return nil, err
	}

	clusters, err := readElasticsearchRemoteClusters(remotes.Resources)
	if err != nil {
		return nil, err
	}
	es.RemoteCluster = clusters

	extensions, err := readElasticsearchExtensions(plan.Elasticsearch)
	if err != nil {
		return nil, err
	}
	es.Extension = extensions

	accounts, err := readElasticsearchTrustAccounts(in.Info.Settings)
	if err != nil {
		return nil, err
	}
	es.TrustAccount = accounts

	externals, err := readElasticsearchTrustExternals(in.Info.Settings)
	if err != nil {
		return nil, err
	}
	es.TrustExternal = externals

	return &es, nil
}

func (es *Elasticsearch) setTopology(topologies ElasticsearchTopologies) {
	set := topologies.AsSet()

	for id, topology := range set {
		topology := topology
		switch id {
		case "hot_content":
			es.HotTier = &topology
		case "coordinating":
			es.CoordinatingTier = &topology
		case "master":
			es.MasterTier = &topology
		case "warm":
			es.WarmTier = &topology
		case "cold":
			es.ColdTier = &topology
		case "frozen":
			es.FrozenTier = &topology
		case "ml":
			es.MlTier = &topology
		}
	}
}

// IsElasticsearchStopped returns true if the resource is stopped.
func IsElasticsearchStopped(res *models.ElasticsearchResourceInfo) bool {
	return res == nil || res.Info == nil || res.Info.Status == nil ||
		*res.Info.Status == "stopped"
}
