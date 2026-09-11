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
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	DefaultReadAfterMutateTimeout  = 30 * time.Second
	DefaultReadAfterMutateInterval = 5 * time.Second
)

// ReadAfterMutate configures GET retries after create/update.
// A nil Timeout is the 30s default; Timeout 0 disables retry.
type ReadAfterMutate struct {
	Wait     func(context.Context, time.Duration)
	Timeout  *time.Duration
	Interval time.Duration
}

func NoRetry() ReadAfterMutate {
	z := time.Duration(0)
	return ReadAfterMutate{Timeout: &z}
}

func ImmediateRetry(timeout time.Duration) ReadAfterMutate {
	t := timeout
	return ReadAfterMutate{
		Wait:    func(context.Context, time.Duration) {},
		Timeout: &t,
	}
}

func (c ReadAfterMutate) wait() func(context.Context, time.Duration) {
	if c.Wait != nil {
		return c.Wait
	}
	return ContextualSleep
}

func (c ReadAfterMutate) timeout() time.Duration {
	if c.Timeout != nil {
		return *c.Timeout
	}
	return DefaultReadAfterMutateTimeout
}

func (c ReadAfterMutate) interval() time.Duration {
	if c.Interval > 0 {
		return c.Interval
	}
	return DefaultReadAfterMutateInterval
}

func (c ReadAfterMutate) UntilFound(
	ctx context.Context,
	get func(context.Context) (found bool, diags diag.Diagnostics),
) (found bool, diags diag.Diagnostics) {
	return UntilFound(ctx, c.wait(), c.timeout(), c.interval(), get)
}

func ContextualSleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

// UntilFound retries get while found is false and diags are empty.
// timeout 0 is a single GET. DeadlineExceeded is not-found; Canceled is an
// error. Callers must check HasError before treating !found as missing.
func UntilFound(
	ctx context.Context,
	wait func(context.Context, time.Duration),
	timeout, interval time.Duration,
	get func(context.Context) (found bool, diags diag.Diagnostics),
) (found bool, diags diag.Diagnostics) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for {
		if diags, stop := stoppedByContext(ctx); stop {
			return false, diags
		}
		found, diags = get(ctx)
		if diags.HasError() {
			return found, diags
		}
		if found {
			if diags, stop := stoppedByContext(ctx); stop {
				return false, diags
			}
			return true, diags
		}
		if timeout == 0 {
			return false, diags
		}
		wait(ctx, interval)
	}
}

func stoppedByContext(ctx context.Context) (diags diag.Diagnostics, stop bool) {
	err := ctx.Err()
	if err == nil {
		return nil, false
	}
	if errors.Is(err, context.Canceled) {
		var d diag.Diagnostics
		d.AddError(err.Error(), err.Error())
		return d, true
	}
	return nil, true
}
