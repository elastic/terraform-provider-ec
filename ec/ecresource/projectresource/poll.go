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

package projectresource

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	// projectInitPollInterval is the interval between project status polls
	// while waiting for a newly created project to become initialised. Project
	// initialisation takes on the order of seconds to minutes, so a 5s interval
	// avoids burning Serverless API rate-limit quota unnecessarily.
	projectInitPollInterval = 5 * time.Second
)

// projectInitPollTimeout bounds the total time spent waiting for a project to
// initialise. It is a var so tests can shorten it. Without a bound a stuck
// project would be polled forever.
var projectInitPollTimeout = 30 * time.Minute

// waitForProjectInitialised polls the project status endpoint until the project
// is initialised or projectInitPollTimeout elapses. 429 responses from the
// Serverless API are retried at the HTTP layer; any other error from getStatus
// is returned immediately.
//
// wait is used to pause between polls and is injected so tests can avoid real
// sleeps. It should return early when ctx is cancelled; the default
// implementation (contextualSleep) does so.
func waitForProjectInitialised(
	ctx context.Context,
	wait func(ctx context.Context, d time.Duration),
	getStatus func(ctx context.Context, id string) (serverless.ProjectStatusPhase, error),
	id string,
) diag.Diagnostics {
	ctx, cancel := context.WithTimeout(ctx, projectInitPollTimeout)
	defer cancel()

	for {
		phase, err := getStatus(ctx, id)
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(err.Error(), err.Error()),
			}
		}
		if phase == serverless.ProjectStatusPhaseInitialized {
			return nil
		}

		if ctx.Err() != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Timed out waiting for project to initialise",
					fmt.Sprintf("Project %s did not reach the initialised phase within %s.", id, projectInitPollTimeout),
				),
			}
		}
		wait(ctx, projectInitPollInterval)
	}
}

// contextualSleep pauses for d, returning early when ctx is cancelled.
//
// It is a var so tests can replace it with a no-op to avoid real sleeps.
var contextualSleep = func(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
