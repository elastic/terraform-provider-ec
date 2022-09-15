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
	"context"
	"flag"
	"log"

	"github.com/elastic/terraform-provider-ec/ec"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

//go:generate go run ./gen/gen.go

// ProviderAddr contains the full name for this terraform provider.
const ProviderAddr = "registry.terraform.io/elastic/ec"

func main() {
	debugFlag := flag.Bool("debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	upgradedSdkProvider, err := tf5to6server.UpgradeServer(
		context.Background(),
		ec.LegacyProvider().GRPCProvider,
	)

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer { return upgradedSdkProvider },
		providerserver.NewProtocol6(ec.New(ec.Version)),
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt

	if *debugFlag {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve(
		ProviderAddr,
		muxServer.ProviderServer,
		serveOpts...,
	)

	if err != nil {
		log.Fatal(err)
	}
}
