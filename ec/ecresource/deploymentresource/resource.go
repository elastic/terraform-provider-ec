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
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Resource returns the ec_deployment resource schema.
func Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: create,
		ReadContext:   read,
		UpdateContext: update,
		DeleteContext: delete,

		Schema: newSchema(),

		Description: "Elastic Cloud Deployment resource",
		Importer: &schema.ResourceImporter{
			// It might be desired to provide the ability to import a deployment
			// specifying key:value pairs of secrets to populate as part of the
			// import with an implementation of schema.StateContextFunc.
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(40 * time.Minute),
			Update:  schema.DefaultTimeout(60 * time.Minute),
			Delete:  schema.DefaultTimeout(60 * time.Minute),
		},
	}
}
