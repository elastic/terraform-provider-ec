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

package serverlesshttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-ec/ec/internal/serverlesshttp"
	"github.com/stretchr/testify/require"
)

// newCountingServer returns a test server whose handler increments a counter
// per request and returns the scripted status sequence, one per request. If
// the sequence is exhausted the handler returns 200.
func newCountingServer(t *testing.T, statuses ...int) (*httptest.Server, *int32) {
	t.Helper()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		idx := int(n) - 1
		status := http.StatusOK
		if idx < len(statuses) {
			status = statuses[idx]
		}
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func TestRetryTransport_Retries429ThenSucceeds(t *testing.T) {
	srv, calls := newCountingServer(t, http.StatusTooManyRequests, http.StatusTooManyRequests)
	client := &http.Client{Transport: serverlesshttp.New(
		serverlesshttp.WithMaxAttempts(5),
		serverlesshttp.WithBaseBackoff(time.Millisecond),
		serverlesshttp.WithMaxBackoff(time.Millisecond),
	)}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
	require.Equal(t, int32(3), atomic.LoadInt32(calls))
}

func TestRetryTransport_StopsAtMaxAttempts(t *testing.T) {
	srv, calls := newCountingServer(t,
		http.StatusTooManyRequests,
		http.StatusTooManyRequests,
		http.StatusTooManyRequests,
		http.StatusTooManyRequests,
	)
	client := &http.Client{Transport: serverlesshttp.New(
		serverlesshttp.WithMaxAttempts(4), // 1 initial + 3 retries
		serverlesshttp.WithBaseBackoff(time.Millisecond),
		serverlesshttp.WithMaxBackoff(time.Millisecond),
	)}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	_ = resp.Body.Close()
	require.Equal(t, int32(4), atomic.LoadInt32(calls))
}

func TestRetryTransport_DoesNotRetryNon429(t *testing.T) {
	srv, calls := newCountingServer(t, http.StatusInternalServerError)
	client := &http.Client{Transport: serverlesshttp.New(
		serverlesshttp.WithMaxAttempts(5),
		serverlesshttp.WithBaseBackoff(time.Millisecond),
	)}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	_ = resp.Body.Close()
	require.Equal(t, int32(1), atomic.LoadInt32(calls)) // no retry
}

func TestRetryTransport_RespectsContextCancellation(t *testing.T) {
	srv, calls := newCountingServer(t,
		http.StatusTooManyRequests,
		http.StatusTooManyRequests,
		http.StatusTooManyRequests,
	)
	client := &http.Client{Transport: serverlesshttp.New(
		serverlesshttp.WithMaxAttempts(10),
		serverlesshttp.WithBaseBackoff(5*time.Second), // long enough to cancel during
		serverlesshttp.WithMaxBackoff(5*time.Second),
	)}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	_, err = client.Do(req)
	require.Error(t, err)
	// Only the initial attempt should have happened before the cancellation
	// kicked in during the first backoff.
	require.Equal(t, int32(1), atomic.LoadInt32(calls))
}

func TestRetryTransport_DisabledWhenMaxAttemptsOne(t *testing.T) {
	srv, calls := newCountingServer(t, http.StatusTooManyRequests)
	client := &http.Client{Transport: serverlesshttp.New(
		serverlesshttp.WithMaxAttempts(1),
		serverlesshttp.WithBaseBackoff(time.Millisecond),
	)}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	_ = resp.Body.Close()
	require.Equal(t, int32(1), atomic.LoadInt32(calls))
}
