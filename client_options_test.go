// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

func TestClientWithCustomOptions(t *testing.T) {
	_, server := createHttpServer(t, index_V3)

	sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
	c, err := NewClient(
		WithSourceURL(sourceURL),
		WithLeveledLogger(new(testleveledLogger)),
		WithLimiter(new(testRateLimiter)),
		WithLogger(new(testLogger)),
		WithRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			return true, nil
		}),
		WithRetryMax(1),
		WithRetryWaitMinMax(100*time.Millisecond, 400*time.Millisecond),
		WithErrorHandler(func(resp *http.Response, err error, numTries int) (*http.Response, error) {
			return resp, nil
		}),
		WithHTTPClient(&http.Client{}),
		WithRequestLogHook(func(logger retryablehttp.Logger, request *http.Request, i int) {
		}),
		WithResponseLogHook(func(logger retryablehttp.Logger, response *http.Response) {
		}),
		WithoutRetries())

	require.NotNil(t, c)
	require.NoError(t, err)
	require.IsType(t, new(testLogger), c.client.Logger)
	require.IsType(t, new(testRateLimiter), c.limiter)
	require.Equal(t, c.client.RetryMax, 1)
	require.Equal(t, c.client.RetryWaitMin, 100*time.Millisecond)
	require.Equal(t, c.client.RetryWaitMax, 400*time.Millisecond)
	require.Equal(t, &http.Client{}, c.client.HTTPClient)
	require.True(t, c.disableRetries)
}

type testLogger struct {
}

func (t testLogger) Printf(string, ...interface{}) {

}

type testRateLimiter struct {
}

func (t *testRateLimiter) Wait(context.Context) error {
	return nil
}

type testleveledLogger struct {
}

func (t *testleveledLogger) Error(msg string, keysAndValues ...interface{}) {
	fmt.Println("Error:"+msg, keysAndValues)
}

func (t *testleveledLogger) Info(msg string, keysAndValues ...interface{}) {
	fmt.Println("Info:"+msg, keysAndValues)
}

func (t *testleveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	fmt.Println("Debug:"+msg, keysAndValues)
}

func (t *testleveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	fmt.Println("Warn:"+msg, keysAndValues)
}
