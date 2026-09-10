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

package util

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

func TestUntilFound_FoundOnFirstTry(t *testing.T) {
	calls := 0
	waits := 0
	wait := func(context.Context, time.Duration) { waits++ }
	get := func(context.Context) (bool, diag.Diagnostics) {
		calls++
		return true, nil
	}

	found, diags := UntilFound(context.Background(), wait, time.Second, time.Millisecond, get)
	require.True(t, found)
	require.False(t, diags.HasError())
	require.Equal(t, 1, calls)
	require.Equal(t, 0, waits)
}

func TestUntilFound_NotFoundThenFound(t *testing.T) {
	calls := 0
	get := func(context.Context) (bool, diag.Diagnostics) {
		calls++
		return calls >= 2, nil
	}

	found, diags := UntilFound(context.Background(), func(context.Context, time.Duration) {}, time.Second, time.Millisecond, get)
	require.True(t, found)
	require.False(t, diags.HasError())
	require.Equal(t, 2, calls)
}

func TestUntilFound_ErrorDiagsStopImmediately(t *testing.T) {
	calls := 0
	waits := 0
	wait := func(context.Context, time.Duration) { waits++ }
	get := func(context.Context) (bool, diag.Diagnostics) {
		calls++
		var d diag.Diagnostics
		d.AddError("boom", "boom")
		return false, d
	}

	found, diags := UntilFound(context.Background(), wait, time.Second, time.Millisecond, get)
	require.False(t, found)
	require.True(t, diags.HasError())
	require.Equal(t, 1, calls)
	require.Equal(t, 0, waits)
}

func TestUntilFound_ErrorAfterMissStops(t *testing.T) {
	calls := 0
	waits := 0
	wait := func(context.Context, time.Duration) { waits++ }
	get := func(context.Context) (bool, diag.Diagnostics) {
		calls++
		if calls == 1 {
			return false, nil
		}
		var d diag.Diagnostics
		d.AddError("boom", "boom")
		return true, d
	}

	found, diags := UntilFound(context.Background(), wait, time.Second, time.Millisecond, get)
	require.True(t, found)
	require.True(t, diags.HasError())
	require.Equal(t, 2, calls)
	require.Equal(t, 1, waits)
}

func TestUntilFound_TimeoutZeroNotFound(t *testing.T) {
	calls := 0
	waits := 0
	wait := func(context.Context, time.Duration) { waits++ }
	get := func(context.Context) (bool, diag.Diagnostics) {
		calls++
		return false, nil
	}

	found, diags := UntilFound(context.Background(), wait, 0, time.Second, get)
	require.False(t, found)
	require.False(t, diags.HasError())
	require.Equal(t, 1, calls)
	require.Equal(t, 0, waits)
}

func TestUntilFound_TimeoutElapsedStillMissing(t *testing.T) {
	calls := 0
	get := func(context.Context) (bool, diag.Diagnostics) {
		calls++
		return false, nil
	}

	done := make(chan struct{})
	var found bool
	var diags diag.Diagnostics
	go func() {
		found, diags = UntilFound(context.Background(), func(context.Context, time.Duration) {}, 50*time.Millisecond, time.Millisecond, get)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("retry loop did not stop")
	}
	require.False(t, found)
	require.Empty(t, diags)
	require.Greater(t, calls, 1)
}

func TestUntilFound_CanceledIsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wait := func(context.Context, time.Duration) { cancel() }
	get := func(context.Context) (bool, diag.Diagnostics) {
		return false, nil
	}

	found, diags := UntilFound(ctx, wait, time.Second, time.Millisecond, get)
	require.False(t, found)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Summary(), "canceled")
}

func TestUntilFound_AlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0

	found, diags := UntilFound(ctx, func(context.Context, time.Duration) {}, time.Second, time.Millisecond, func(context.Context) (bool, diag.Diagnostics) {
		calls++
		return false, nil
	})
	require.False(t, found)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Summary(), "canceled")
	require.Equal(t, 0, calls)
}

func TestUntilFound_CanceledAfterFound(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	get := func(context.Context) (bool, diag.Diagnostics) {
		cancel()
		return true, nil
	}

	found, diags := UntilFound(ctx, func(context.Context, time.Duration) {}, time.Second, time.Millisecond, get)
	require.False(t, found)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Summary(), "canceled")
}

func TestReadAfterMutate_ZeroValueUsesDefaults(t *testing.T) {
	var c ReadAfterMutate
	require.Equal(t, DefaultReadAfterMutateTimeout, c.timeout())
	require.Equal(t, DefaultReadAfterMutateInterval, c.interval())
	require.NotNil(t, c.wait())
}

func TestReadAfterMutate_NoRetryDisablesRetry(t *testing.T) {
	require.Equal(t, time.Duration(0), NoRetry().timeout())
}

func TestReadAfterMutate_UntilFoundUsesConfig(t *testing.T) {
	calls := 0
	found, diags := ImmediateRetry(time.Second).UntilFound(context.Background(), func(context.Context) (bool, diag.Diagnostics) {
		calls++
		return calls >= 2, nil
	})
	require.True(t, found)
	require.False(t, diags.HasError())
	require.Equal(t, 2, calls)
}
