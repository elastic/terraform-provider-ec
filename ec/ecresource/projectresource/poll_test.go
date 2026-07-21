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
	"errors"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/stretchr/testify/require"
)

// withNoopSleep swaps contextualSleep for a no-op for the duration of the test,
// so polling tests don't sleep for real.
func withNoopSleep(t *testing.T) {
	t.Helper()
	orig := contextualSleep
	contextualSleep = func(context.Context, time.Duration) {}
	t.Cleanup(func() { contextualSleep = orig })
}

func TestWaitForProjectInitialised_ReturnsWhenInitialised(t *testing.T) {
	withNoopSleep(t)

	calls := 0
	getStatus := func(_ context.Context, _ string) (serverless.ProjectStatusPhase, error) {
		calls++
		if calls < 3 {
			return serverless.ProjectStatusPhaseInitializing, nil
		}
		return serverless.ProjectStatusPhaseInitialized, nil
	}

	diags := waitForProjectInitialised(context.Background(), contextualSleep, getStatus, "id")
	require.False(t, diags.HasError())
	require.Equal(t, 3, calls)
}

func TestWaitForProjectInitialised_PropagatesGetStatusError(t *testing.T) {
	withNoopSleep(t)

	wantErr := errors.New("boom")
	getStatus := func(_ context.Context, _ string) (serverless.ProjectStatusPhase, error) {
		return "", wantErr
	}

	diags := waitForProjectInitialised(context.Background(), contextualSleep, getStatus, "id")
	require.True(t, diags.HasError())
	require.Equal(t, "boom", diags[0].Summary())
	require.Equal(t, "boom", diags[0].Detail())
}

func TestWaitForProjectInitialised_TimesOut(t *testing.T) {
	// Override the timeout to something tiny so the test is fast.
	origTimeout := projectInitPollTimeout
	projectInitPollTimeout = 50 * time.Millisecond
	t.Cleanup(func() { projectInitPollTimeout = origTimeout })
	withNoopSleep(t)

	getStatus := func(_ context.Context, _ string) (serverless.ProjectStatusPhase, error) {
		return serverless.ProjectStatusPhaseInitializing, nil
	}

	diags := waitForProjectInitialised(context.Background(), contextualSleep, getStatus, "id")
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Summary(), "Timed out waiting for project to initialise")
	require.Contains(t, diags[0].Detail(), "id")
}
