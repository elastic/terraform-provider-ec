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

package main

import (
	"flag"

	"github.com/elastic/terraform-provider-ec/ec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

//go:generate go run ./gen/gen.go

// ProviderAddr contains the full name for this terraform provider.
const ProviderAddr = "registry.terraform.io/elastic/ec"

func main() {
	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ec.Provider,
		Debug:        debugMode,
		ProviderAddr: ProviderAddr,
	})
}
