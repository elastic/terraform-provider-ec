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
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

// flattenEsResources takes in Elasticsearch resource models and returns its
// flattened form.
func flattenEsResources(in []*models.ElasticsearchResourceInfo, name string, remotes models.RemoteResources) ([]interface{}, error) {
	result := make([]interface{}, 0, len(in))
	for _, res := range in {
		m := make(map[string]interface{})
		if util.IsCurrentEsPlanEmpty(res) || isEsResourceStopped(res) {
			continue
		}

		if res.Info.ClusterID != nil && *res.Info.ClusterID != "" {
			m["resource_id"] = *res.Info.ClusterID
		}

		if res.RefID != nil && *res.RefID != "" {
			m["ref_id"] = *res.RefID
		}

		if res.Region != nil {
			m["region"] = *res.Region
		}

		plan := res.Info.PlanInfo.Current.Plan
		topology, err := flattenEsTopology(plan)
		if err != nil {
			return nil, err
		}
		if len(topology) > 0 {
			m["topology"] = topology
		}

		if plan.AutoscalingEnabled != nil {
			m["autoscale"] = strconv.FormatBool(*plan.AutoscalingEnabled)
		}

		if meta := res.Info.Metadata; meta != nil && meta.CloudID != "" {
			m["cloud_id"] = meta.CloudID
		}

		for k, v := range util.FlattenClusterEndpoint(res.Info.Metadata) {
			m[k] = v
		}

		m["config"] = flattenEsConfig(plan.Elasticsearch)

		if remotes := flattenEsRemotes(remotes); remotes.Len() > 0 {
			m["remote_cluster"] = remotes
		}

		extensions := schema.NewSet(esExtensionHash, nil)
		for _, ext := range flattenEsBundles(plan.Elasticsearch.UserBundles) {
			extensions.Add(ext)
		}

		for _, ext := range flattenEsPlugins(plan.Elasticsearch.UserPlugins) {
			extensions.Add(ext)
		}

		if extensions.Len() > 0 {
			m["extension"] = extensions
		}

		if settings := res.Info.Settings; settings != nil {
			if trust := flattenAccountTrust(settings.Trust); trust != nil {
				m["trust_account"] = trust
			}

			if trust := flattenExternalTrust(settings.Trust); trust != nil {
				m["trust_external"] = trust
			}
		}

		result = append(result, m)
	}

	return result, nil
}

func flattenEsTopology(plan *models.ElasticsearchClusterPlan) ([]interface{}, error) {
	result := make([]interface{}, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var m = make(map[string]interface{})
		if topology.Size == nil || topology.Size.Value == nil || *topology.Size.Value == 0 {
			continue
		}

		// ID is always set.
		m["id"] = topology.ID

		if topology.InstanceConfigurationID != "" {
			m["instance_configuration_id"] = topology.InstanceConfigurationID
		}

		// TODO: Check legacy plans.
		// if topology.MemoryPerNode > 0 {
		// 	m["size"] = strconv.Itoa(int(topology.MemoryPerNode))
		// }

		if topology.Size != nil {
			m["size"] = util.MemoryToState(*topology.Size.Value)
			m["size_resource"] = *topology.Size.Resource
		}

		m["zone_count"] = topology.ZoneCount

		if nt := topology.NodeType; nt != nil {
			if nt.Data != nil {
				m["node_type_data"] = strconv.FormatBool(*nt.Data)
			}

			if nt.Ingest != nil {
				m["node_type_ingest"] = strconv.FormatBool(*nt.Ingest)
			}

			if nt.Master != nil {
				m["node_type_master"] = strconv.FormatBool(*nt.Master)
			}

			if nt.Ml != nil {
				m["node_type_ml"] = strconv.FormatBool(*nt.Ml)
			}
		}

		if len(topology.NodeRoles) > 0 {
			m["node_roles"] = schema.NewSet(schema.HashString, util.StringToItems(
				topology.NodeRoles...,
			))
		}

		autoscaling := make(map[string]interface{})
		if ascale := topology.AutoscalingMax; ascale != nil {
			autoscaling["max_size_resource"] = *ascale.Resource
			autoscaling["max_size"] = util.MemoryToState(*ascale.Value)
		}

		if ascale := topology.AutoscalingMin; ascale != nil {
			autoscaling["min_size_resource"] = *ascale.Resource
			autoscaling["min_size"] = util.MemoryToState(*ascale.Value)
		}

		if topology.AutoscalingPolicyOverrideJSON != nil {
			b, err := json.Marshal(topology.AutoscalingPolicyOverrideJSON)
			if err != nil {
				return nil, fmt.Errorf(
					"elasticsearch topology %s: unable to persist policy_override_json: %w",
					topology.ID, err,
				)
			}
			autoscaling["policy_override_json"] = string(b)
		}

		if len(autoscaling) > 0 {
			m["autoscaling"] = []interface{}{autoscaling}
		}

		// Computed config object to avoid unsetting legacy topology config settings.
		m["config"] = flattenEsConfig(topology.Elasticsearch)

		result = append(result, m)
	}

	// Ensure the topologies are sorted alphabetically by ID.
	sort.SliceStable(result, func(i, j int) bool {
		a := result[i].(map[string]interface{})
		b := result[j].(map[string]interface{})
		return a["id"].(string) < b["id"].(string)
	})
	return result, nil
}

func flattenEsConfig(cfg *models.ElasticsearchConfiguration) []interface{} {
	var m = make(map[string]interface{})
	if cfg == nil {
		return nil
	}

	if len(cfg.EnabledBuiltInPlugins) > 0 {
		m["plugins"] = schema.NewSet(schema.HashString,
			util.StringToItems(cfg.EnabledBuiltInPlugins...),
		)
	}

	if cfg.UserSettingsYaml != "" {
		m["user_settings_yaml"] = cfg.UserSettingsYaml
	}

	if cfg.UserSettingsOverrideYaml != "" {
		m["user_settings_override_yaml"] = cfg.UserSettingsOverrideYaml
	}

	if o := cfg.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			m["user_settings_json"] = string(b)
		}
	}

	if o := cfg.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			m["user_settings_override_json"] = string(b)
		}
	}

	if cfg.DockerImage != "" {
		m["docker_image"] = cfg.DockerImage
	}

	// If no settings are set, there's no need to store the empty values in the
	// state and makes the state consistent with a clean import return.
	if len(m) == 0 {
		return nil
	}

	return []interface{}{m}
}

func flattenEsRemotes(in models.RemoteResources) *schema.Set {
	res := newElasticsearchRemoteSet()
	for _, r := range in.Resources {
		var m = make(map[string]interface{})
		if r.DeploymentID != nil && *r.DeploymentID != "" {
			m["deployment_id"] = *r.DeploymentID
		}

		if r.ElasticsearchRefID != nil && *r.ElasticsearchRefID != "" {
			m["ref_id"] = *r.ElasticsearchRefID
		}

		if r.Alias != nil && *r.Alias != "" {
			m["alias"] = *r.Alias
		}

		if r.SkipUnavailable != nil {
			m["skip_unavailable"] = *r.SkipUnavailable
		}
		res.Add(m)
	}

	return res
}

func newElasticsearchRemoteSet(remotes ...interface{}) *schema.Set {
	return schema.NewSet(
		schema.HashResource(elasticsearchRemoteCluster().Elem.(*schema.Resource)),
		remotes,
	)
}

func flattenEsBundles(in []*models.ElasticsearchUserBundle) []interface{} {
	result := make([]interface{}, 0, len(in))
	for _, bundle := range in {
		m := make(map[string]interface{})
		m["type"] = "bundle"
		m["version"] = *bundle.ElasticsearchVersion
		m["url"] = *bundle.URL
		m["name"] = *bundle.Name

		result = append(result, m)
	}

	return result
}

func flattenEsPlugins(in []*models.ElasticsearchUserPlugin) []interface{} {
	result := make([]interface{}, 0, len(in))
	for _, plugin := range in {
		m := make(map[string]interface{})
		m["type"] = "plugin"
		m["version"] = *plugin.ElasticsearchVersion
		m["url"] = *plugin.URL
		m["name"] = *plugin.Name

		result = append(result, m)
	}

	return result
}

func flattenAccountTrust(in *models.ElasticsearchClusterTrustSettings) *schema.Set {
	if in == nil {
		return nil
	}

	account := schema.NewSet(schema.HashResource(accountResource()), nil)
	for _, acc := range in.Accounts {
		account.Add(map[string]interface{}{
			"account_id": *acc.AccountID,
			"trust_all":  *acc.TrustAll,
			"trust_allowlist": schema.NewSet(schema.HashString,
				util.StringToItems(acc.TrustAllowlist...),
			),
		})
	}

	if account.Len() > 0 {
		return account
	}
	return nil
}

func flattenExternalTrust(in *models.ElasticsearchClusterTrustSettings) *schema.Set {
	if in == nil {
		return nil
	}

	external := schema.NewSet(schema.HashResource(externalResource()), nil)
	for _, ext := range in.External {
		external.Add(map[string]interface{}{
			"relationship_id": *ext.TrustRelationshipID,
			"trust_all":       *ext.TrustAll,
			"trust_allowlist": schema.NewSet(schema.HashString,
				util.StringToItems(ext.TrustAllowlist...),
			),
		})
	}

	if external.Len() > 0 {
		return external
	}
	return nil
}
