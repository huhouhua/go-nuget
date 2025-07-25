// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWithHeader(t *testing.T) {
	mux, client := setup(t, index_V3)
	mux.HandleFunc("/v3/without-header", func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.Header.Get("X-CUSTOM-HEADER"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"X-CUSTOM-HEADER": %s`, r.Header.Get("X-CUSTOM-HEADER"))
	})
	mux.HandleFunc("/v3/with-header", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "randomtokenstring", r.Header.Get("X-CUSTOM-HEADER"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"X-CUSTOM-HEADER": %s`, r.Header.Get("X-CUSTOM-HEADER"))
	})

	// ensure that X-CUSTOM-HEADER hasn't been set at all
	req, err := client.NewRequest(http.MethodGet, "/v3/without-header", nil, nil, nil)
	require.NoError(t, err)

	_, err = client.Do(req, nil, DecoderEmpty)
	require.NoError(t, err)

	// ensure that X-CUSTOM-HEADER is set for only one request
	req, err = client.NewRequest(
		http.MethodGet,
		"/v3/with-header",
		nil,
		nil,
		[]RequestOptionFunc{WithHeader("X-CUSTOM-HEADER", "randomtokenstring")},
	)
	require.NoError(t, err)

	_, err = client.Do(req, nil, DecoderEmpty)
	require.NoError(t, err)

	req, err = client.NewRequest(http.MethodGet, "/v3/without-header", nil, nil, nil)
	require.NoError(t, err)

	_, err = client.Do(req, nil, DecoderEmpty)
	require.NoError(t, err)

	client, err = NewClient(
		// same base options as setup
		WithSourceURL(client.SourceURL().String()),
		// Disable backoff to speed up tests that expect errors.
		WithBackoff(func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
			return 0
		}),
		// add client headers
		WithRequestOptions(WithHeader("X-CUSTOM-HEADER", "randomtokenstring")))
	require.NoError(t, err)
	require.NotNil(t, client)

	req, err = client.NewRequest(http.MethodGet, "/v3/with-header", nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, "randomtokenstring", req.Header.Get("X-CUSTOM-HEADER"))

	_, err = client.Do(req, nil, DecoderEmpty)
	require.NoError(t, err)

	req, err = client.NewRequest(http.MethodGet, "/v3/with-header", nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, "randomtokenstring", req.Header.Get("X-CUSTOM-HEADER"))

	_, err = client.Do(req, nil, DecoderEmpty)
	require.NoError(t, err)
}

func TestWithHeaders(t *testing.T) {
	mux, client := setup(t, index_V3)
	mux.HandleFunc("/v3/without-headers", func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.Header.Get("X-CUSTOM-HEADER-1"))
		require.Empty(t, r.Header.Get("X-CUSTOM-HEADER-2"))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v3/with-headers", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "randomtokenstring", r.Header.Get("X-CUSTOM-HEADER-1"))
		require.Equal(t, "randomtokenstring2", r.Header.Get("X-CUSTOM-HEADER-2"))
		w.WriteHeader(http.StatusOK)
	})

	headers := map[string]string{
		"X-CUSTOM-HEADER-1": "randomtokenstring",
		"X-CUSTOM-HEADER-2": "randomtokenstring2",
	}

	// ensure that X-CUSTOM-HEADER hasn't been set at all
	req, err := client.NewRequest(http.MethodGet, "/v3/without-headers", nil, nil, nil)
	require.NoError(t, err)

	_, err = client.Do(req, nil, DecoderEmpty)
	require.NoError(t, err)

	// ensure that X-CUSTOM-HEADER is set for only one request
	req, err = client.NewRequest(
		http.MethodGet,
		"/v3/with-headers",
		nil,
		nil,
		[]RequestOptionFunc{WithHeaders(headers)},
	)
	require.NoError(t, err)

	_, err = client.Do(req, nil, DecoderEmpty)
	require.NoError(t, err)

	req, err = client.NewRequest(http.MethodGet, "/v3/without-headers", nil, nil, nil)
	require.NoError(t, err)

	_, err = client.Do(req, nil, DecoderEmpty)
	require.NoError(t, err)

	client, err = NewClient(
		// same base options as setup
		WithSourceURL(client.SourceURL().String()),
		// Disable backoff to speed up tests that expect errors.
		WithBackoff(func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
			return 0
		}),
		// add client headers
		WithRequestOptions(WithHeaders(headers)),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	req, err = client.NewRequest(http.MethodGet, "/v3/with-headers", nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, "randomtokenstring", req.Header.Get("X-CUSTOM-HEADER-1"))

	_, err = client.Do(req, nil, DecoderEmpty)
	require.NoError(t, err)

	req, err = client.NewRequest(http.MethodGet, "/v3/with-headers", nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, "randomtokenstring", req.Header.Get("X-CUSTOM-HEADER-1"))

	_, err = client.Do(req, nil, DecoderEmpty)
	require.NoError(t, err)
}

func TestWithHeaderNuget(t *testing.T) {
	_, client := setup(t, index_V3)
	req, err := client.NewRequest(
		http.MethodGet,
		"/v3/with-nuget",
		nil,
		nil,
		[]RequestOptionFunc{WithAPIKey("test-with-api-key"), WithNugetClientVersion("5.0.0")},
	)
	require.NoError(t, err)
	require.NotNil(t, req)
	require.Equal(t, "test-with-api-key", req.Header.Get("X-NuGet-ApiKey"))
	require.Equal(t, "5.0.0", req.Header.Get("X-NuGet-Client-Version"))
}
