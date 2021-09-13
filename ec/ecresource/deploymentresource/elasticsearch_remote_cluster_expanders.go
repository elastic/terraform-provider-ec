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
	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func handleRemoteClusters(d *schema.ResourceData, client *api.API) error {
	if keyIsEmptyUnchanged(d, "elasticsearch.0.remote_cluster") {
		return nil
	}

	remoteResources := expandRemoteClusters(
		d.Get("elasticsearch.0.remote_cluster").(*schema.Set),
	)

	return esremoteclustersapi.Update(esremoteclustersapi.UpdateParams{
		API:             client,
		DeploymentID:    d.Id(),
		RefID:           d.Get("elasticsearch.0.ref_id").(string),
		RemoteResources: remoteResources,
	})
}

func expandRemoteClusters(set *schema.Set) *models.RemoteResources {
	res := models.RemoteResources{Resources: []*models.RemoteResourceRef{}}

	for _, r := range set.List() {
		var resourceRef models.RemoteResourceRef
		m := r.(map[string]interface{})

		if id, ok := m["deployment_id"]; ok {
			resourceRef.DeploymentID = ec.String(id.(string))
		}

		if v, ok := m["ref_id"]; ok {
			resourceRef.ElasticsearchRefID = ec.String(v.(string))
		}

		if v, ok := m["alias"]; ok {
			resourceRef.Alias = ec.String(v.(string))
		}

		if v, ok := m["skip_unavailable"]; ok {
			resourceRef.SkipUnavailable = ec.Bool(v.(bool))
		}

		res.Resources = append(res.Resources, &resourceRef)
	}

	return &res
}

func keyIsEmptyUnchanged(d *schema.ResourceData, k string) bool {
	old, new := d.GetChange(k)
	oldSlice := old.(*schema.Set)
	newSlice := new.(*schema.Set)
	return oldSlice.Len() == 0 && newSlice.Len() == 0
}
