// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"context"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// RequestOptionFunc can be passed to all API requests to customize the API request.
type RequestOptionFunc func(*retryablehttp.Request) error

// WithContext runs the request with the provided context
func WithContext(ctx context.Context) RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		*req = *req.WithContext(ctx)
		return nil
	}
}

// WithHeader takes a header name and value and appends it to the request headers.
func WithHeader(name, value string) RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		req.Header.Set(name, value)
		return nil
	}
}

// WithHeaders takes a map of header name/value pairs and appends them to the
// request headers.
func WithHeaders(headers map[string]string) RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return nil
	}
}

// WithAPIKey takes a apiKey which is then used when making this one request.
func WithAPIKey(apiKey string) RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		req.Header.Set("X-NuGet-ApiKey", apiKey)
		return nil
	}
}
