// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"net/http"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// ClientOptionFunc can be used to customize a new NuGet API client.
type ClientOptionFunc func(*Client) error

// WithSourceURL sets the source URL for API requests to a custom endpoint.
func WithSourceURL(urlStr string) ClientOptionFunc {
	return func(c *Client) error {
		return c.setSourceURL(urlStr)
	}
}

// WithBackoff can be used to configure a custom backoff policy.
func WithBackoff(backoff retryablehttp.Backoff) ClientOptionFunc {
	return func(c *Client) error {
		c.client.Backoff = backoff
		return nil
	}
}

// WithLeveledLogger can be used to configure a custom retryablehttp
// leveled logger.
func WithLeveledLogger(leveledLogger retryablehttp.LeveledLogger) ClientOptionFunc {
	return func(c *Client) error {
		c.client.Logger = leveledLogger
		return nil
	}
}

// WithLimiter injects a custom rate limiter to the client.
func WithLimiter(limiter RateLimiter) ClientOptionFunc {
	return func(c *Client) error {
		c.configureLimiterOnce.Do(func() {})
		c.limiter = limiter
		return nil
	}
}

// WithLogger can be used to configure a custom retryablehttp logger.
func WithLogger(logger retryablehttp.Logger) ClientOptionFunc {
	return func(c *Client) error {
		c.client.Logger = logger
		return nil
	}
}

// WithRetry can be used to configure a custom retry policy.
func WithRetry(checkRetry retryablehttp.CheckRetry) ClientOptionFunc {
	return func(c *Client) error {
		c.client.CheckRetry = checkRetry
		return nil
	}
}

// WithRetryMax can be used to configure a custom maximum number of retries.
func WithRetryMax(retryMax int) ClientOptionFunc {
	return func(c *Client) error {
		c.client.RetryMax = retryMax
		return nil
	}
}

// WithRetryWaitMinMax can be used to configure a custom minimum and
// maximum time to wait between retries.
func WithRetryWaitMinMax(waitMin, waitMax time.Duration) ClientOptionFunc {
	return func(c *Client) error {
		c.client.RetryWaitMin = waitMin
		c.client.RetryWaitMax = waitMax
		return nil
	}
}

// WithErrorHandler can be used to configure a custom error handler.
func WithErrorHandler(handler retryablehttp.ErrorHandler) ClientOptionFunc {
	return func(c *Client) error {
		c.client.ErrorHandler = handler
		return nil
	}
}

// WithHTTPClient can be used to configure a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOptionFunc {
	return func(c *Client) error {
		c.client.HTTPClient = httpClient
		return nil
	}
}

// WithRequestLogHook can be used to configure a custom request log hook.
func WithRequestLogHook(hook retryablehttp.RequestLogHook) ClientOptionFunc {
	return func(c *Client) error {
		c.client.RequestLogHook = hook
		return nil
	}
}

// WithResponseLogHook can be used to configure a custom response log hook.
func WithResponseLogHook(hook retryablehttp.ResponseLogHook) ClientOptionFunc {
	return func(c *Client) error {
		c.client.ResponseLogHook = hook
		return nil
	}
}

// WithoutRetries disables the default retry logic.
func WithoutRetries() ClientOptionFunc {
	return func(c *Client) error {
		c.disableRetries = true
		return nil
	}
}

// WithRequestOptions can be used to configure default request options applied to every request.
func WithRequestOptions(options ...RequestOptionFunc) ClientOptionFunc {
	return func(c *Client) error {
		c.defaultRequestOptions = append(c.defaultRequestOptions, options...)
		return nil
	}
}
