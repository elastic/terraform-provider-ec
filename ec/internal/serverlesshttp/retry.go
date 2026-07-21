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

// Package serverlesshttp provides an http.RoundTripper that retries
// 429 Too Many Requests responses from the Elastic Cloud Serverless API.
//
// Only 429 is retried. The Serverless API does not document a Retry-After
// header, so retries use self-paced exponential backoff with jitter. All other
// responses (including 5xx) are returned to the caller untouched, which avoids
// any idempotency concerns around retrying write methods.
package serverlesshttp

import (
	"math/rand/v2"
	"net/http"
	"time"
)

// RetryTransport retries 429 Too Many Requests responses with bounded
// exponential backoff and jitter.
type RetryTransport struct {
	// Next is the underlying transport. Defaults to http.DefaultTransport.
	Next http.RoundTripper

	// MaxAttempts is the total number of attempts including the first. A
	// value of 1 disables retries. Defaults to 5.
	MaxAttempts int

	// BaseBackoff is the initial backoff duration. Defaults to 1 second.
	BaseBackoff time.Duration

	// MaxBackoff caps the backoff duration. Defaults to 30 seconds.
	MaxBackoff time.Duration
}

// New returns a RetryTransport wrapping the given transport (or
// http.DefaultTransport if next is nil) with sensible defaults.
func New(opts ...Option) *RetryTransport {
	t := &RetryTransport{
		Next:        http.DefaultTransport,
		MaxAttempts: 5,
		BaseBackoff: 1 * time.Second,
		MaxBackoff:  30 * time.Second,
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

// Option configures a RetryTransport.
type Option func(*RetryTransport)

// WithNext sets the underlying transport.
func WithNext(next http.RoundTripper) Option {
	return func(t *RetryTransport) { t.Next = next }
}

// WithMaxAttempts sets the total number of attempts including the first.
func WithMaxAttempts(n int) Option {
	return func(t *RetryTransport) { t.MaxAttempts = n }
}

// WithBaseBackoff sets the initial backoff duration.
func WithBaseBackoff(d time.Duration) Option {
	return func(t *RetryTransport) { t.BaseBackoff = d }
}

// WithMaxBackoff caps the backoff duration.
func WithMaxBackoff(d time.Duration) Option {
	return func(t *RetryTransport) { t.MaxBackoff = d }
}

// RoundTrip executes the request, retrying 429 responses up to MaxAttempts
// times with exponential backoff and jitter.
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Propagate an already-cancelled context immediately.
	if err := req.Context().Err(); err != nil {
		return nil, err
	}

	resp, err := t.next().RoundTrip(req)
	for attempt := 1; attempt < t.maxAttempts() && isTooManyRequests(resp, err); attempt++ {
		// Close the response body so the connection can be reused.
		if resp != nil {
			_ = resp.Body.Close()
		}

		wait := backoff(attempt, t.baseBackoff(), t.maxBackoff())
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(wait):
		}

		resp, err = t.next().RoundTrip(req)
	}
	return resp, err
}

func (t *RetryTransport) next() http.RoundTripper {
	if t.Next == nil {
		return http.DefaultTransport
	}
	return t.Next
}

func (t *RetryTransport) maxAttempts() int {
	if t.MaxAttempts <= 0 {
		return 1
	}
	return t.MaxAttempts
}

func (t *RetryTransport) baseBackoff() time.Duration {
	if t.BaseBackoff <= 0 {
		return 1 * time.Second
	}
	return t.BaseBackoff
}

func (t *RetryTransport) maxBackoff() time.Duration {
	if t.MaxBackoff <= 0 {
		return 30 * time.Second
	}
	return t.MaxBackoff
}

// isTooManyRequests reports whether the result should be retried. Only a 429
// response is retried; transport errors and all other status codes are not.
func isTooManyRequests(resp *http.Response, err error) bool {
	return err == nil && resp != nil && resp.StatusCode == http.StatusTooManyRequests
}

// backoff returns the wait duration for the nth retry (1-indexed):
// base*2^(n-1), capped at max, with up to +50% jitter.
func backoff(n int, base, max time.Duration) time.Duration {
	d := base << (n - 1) // base, 2*base, 4*base, ...
	if d <= 0 || d > max {
		d = max
	}
	jitter := time.Duration(rand.Int64N(int64(d) / 2))
	return d + jitter
}
