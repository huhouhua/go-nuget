// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	index_V3    = "testdata/index.json"
	index_Baget = "testdata/index_baget.json"
)

// setup sets up a test HTTP server along with a NuGet.Client that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup(t *testing.T, indexPath string) (*http.ServeMux, *Client) {
	mux, server := createHttpServer(t, indexPath)

	// client is the NuGet client being tested.
	client, err := NewOAuthClient("test_go_nuget_key",
		WithBaseURL(server.URL),
		// Disable backoff to speed up tests that expect errors.
		WithCustomBackoff(func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
			return 0
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	for _, u := range client.serviceUrls {
		u.Host = client.baseURL.Host
		u.Scheme = client.baseURL.Scheme
	}
	return mux, client
}

func createHttpServer(t *testing.T, indexPath string) (*http.ServeMux, *httptest.Server) {
	// mux is the HTTP request multiplexer used with the test server.
	mux := http.NewServeMux()

	mux.HandleFunc("/v3/index.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, indexPath)
	})

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return mux, server
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %s, want %s", got, want)
	}
}

func mustWriteHTTPResponse(t *testing.T, w io.Writer, fixturePath string) {
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("error opening fixture file: %v", err)
	}

	if _, err = io.Copy(w, f); err != nil {
		t.Fatalf("error writing response: %v", err)
	}
}

// Helper to make absolute path in test cases
func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return abs
}

// Helper function to create a temporary directory and files for testing
func createTestDirectory(t *testing.T, dirName string, files []string) string {
	dirPath := filepath.Join(t.TempDir(), dirName)
	createEmptyDir(t, dirPath)

	// Create the files in the directory
	for _, file := range files {
		filePath := filepath.Join(dirPath, file)
		f, err := os.Create(filePath)
		require.NoErrorf(t, err, "Failed to create file in test directory: %v", err)
		_ = f.Close()
	}

	return dirPath
}

func createFile(t *testing.T, path, data string) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(t, err)

	err = os.WriteFile(path, []byte(data), 0644)
	require.NoError(t, err)
}

func createEmptyDir(t *testing.T, path string) {
	err := os.MkdirAll(path, 0755)
	require.NoError(t, err)
}

func TestNewClient(t *testing.T) {
	_, server := createHttpServer(t, index_V3)
	c, err := NewClient(WithBaseURL(server.URL))

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	for serviceType, u := range c.serviceUrls {
		t.Logf("type:%s url:%s", serviceType.String(), u.String())
	}
	//expectedBaseURL := defaultBaseURL+apiVersionPath

	if c.BaseURL().String() != server.URL {
		t.Errorf("NewClient BaseURL is %s, want %s", c.BaseURL().String(), server.URL)
	}
	if c.UserAgent != userAgent {
		t.Errorf("NewClient UserAgent is %s, want %s", c.UserAgent, userAgent)
	}
}

func TestNewOAuthClient(t *testing.T) {
	t.Run("new a oath client return success", func(t *testing.T) {
		_, server := createHttpServer(t, index_V3)
		c, err := NewOAuthClient("", WithBaseURL(server.URL))

		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		if c.BaseURL().String() != server.URL {
			t.Errorf("NewOAuthClient BaseURL is %s, want %s", c.BaseURL().String(), server.URL)
		}
		if c.UserAgent != userAgent {
			t.Errorf("NewOAuthClient UserAgent is %s, want %s", c.UserAgent, userAgent)
		}
	})
	t.Run("new client return error", func(t *testing.T) {
		wantError := fmt.Errorf("options fail")
		_, err := NewOAuthClient("", nil, func(client *Client) error {
			return wantError
		})
		require.Equal(t, wantError, err)
	})
}

func TestClient_Retry(t *testing.T) {
	mux, server := createHttpServer(t, index_V3)
	c, err := NewClient(WithBaseURL(server.URL))
	require.NoError(t, err)
	c.client.RetryMax = 1

	t.Run("retry http check return error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()

		ok, err := c.retryHTTPCheck(ctx, nil, nil)
		require.False(t, ok)
		require.Equal(t, context.DeadlineExceeded, err)

		r, err := c.NewRequest(http.MethodGet, "", nil, nil, nil)
		require.NoError(t, err)

		r.URL = nil
		_, err = c.Do(r, nil, DecoderEmpty)
		wantErr := &url.Error{
			Op:  "Get",
			Err: errors.New("http: nil Request.URL"),
		}
		require.Equal(t, wantErr, err)
	})
	t.Run("retry http backoff with statusTooMany return error", func(t *testing.T) {
		mux.HandleFunc("/retryHTTPBackoff/statusTooMany", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		})
		r, err := c.NewRequest(http.MethodGet, "retryHTTPBackoff/statusTooMany", nil, nil, nil)
		require.NoError(t, err)

		_, err = c.Do(r, nil, DecoderEmpty)
		respErr := &ErrorResponse{}
		require.True(t, errors.As(err, &respErr))
		require.Equal(t, http.StatusTooManyRequests, respErr.Response.StatusCode)
	})
	t.Run("retry http backoff with statusBadGateway return error", func(t *testing.T) {
		mux.HandleFunc("/retryHTTPBackoff/statusBadGateway", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		})
		r, err := c.NewRequest(http.MethodGet, "retryHTTPBackoff/statusBadGateway", nil, nil, nil)
		require.NoError(t, err)

		_, err = c.Do(r, nil, DecoderEmpty)
		respErr := &ErrorResponse{}
		require.True(t, errors.As(err, &respErr))
		require.Equal(t, http.StatusBadGateway, respErr.Response.StatusCode)
	})
	t.Run("retry http backoff header include rateLimit-reset return error", func(t *testing.T) {
		wantRateReset := strconv.FormatInt(int64(c.client.RetryWaitMax*time.Millisecond+1), 10)

		mux.HandleFunc("/retryHTTPBackoff/header/statusTooMany", func(w http.ResponseWriter, r *http.Request) {
			header := w.Header()
			header.Add(headerRateReset, wantRateReset)
			w.WriteHeader(http.StatusTooManyRequests)
		})

		r, err := c.NewRequest(http.MethodGet, "retryHTTPBackoff/header/statusTooMany", nil, nil, nil)
		require.NoError(t, err)

		resp, err := c.Do(r, nil, DecoderEmpty)
		respErr := &ErrorResponse{}
		require.True(t, errors.As(err, &respErr))
		require.Equal(t, http.StatusTooManyRequests, respErr.Response.StatusCode)

		require.Equal(t, wantRateReset, resp.Header.Get(headerRateReset))
	})
}

func TestClient_configureLimiter(t *testing.T) {
	mux, server := createHttpServer(t, index_V3)
	c, err := NewClient(WithBaseURL(server.URL))
	require.NoError(t, err)

	c.configureLimiterOnce = sync.Once{}
	wantRateLimit := strconv.FormatInt(int64(c.client.RetryWaitMax*time.Millisecond+1), 10)

	mux.HandleFunc("/configureLimiter", func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Add(headerRateLimit, wantRateLimit)
		w.WriteHeader(http.StatusOK)
	})
	r, err := c.NewRequest(http.MethodGet, "configureLimiter", nil, nil, nil)
	require.NoError(t, err)

	resp, err := c.Do(r, nil, DecoderEmpty)
	require.NoError(t, err)
	require.Equal(t, wantRateLimit, resp.Header.Get(headerRateLimit))

}

func TestCheckResponseOnHeadRequestError(t *testing.T) {
	_, server := createHttpServer(t, index_V3)
	c, err := NewClient(WithBaseURL(server.URL))

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	req, err := c.NewRequest(http.MethodHead, "test", nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp := &http.Response{
		Request:    req.Request,
		StatusCode: http.StatusNotFound,
		Body:       nil,
	}

	errResp := CheckResponse(resp)
	if errResp == nil {
		t.Fatal("Expected error response.")
	}

	want := "404 Not Found"

	if errResp.Error() != want {
		t.Errorf("Expected error: %s, got %s", want, errResp.Error())
	}
}

func TestRequestWithContext(t *testing.T) {
	_, server := createHttpServer(t, index_V3)
	c, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	req, err := c.NewRequest(http.MethodGet, "test", nil, nil, []RequestOptionFunc{WithContext(ctx)})
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	t.Cleanup(func() {
		cancel()
	})

	if req.Context() != ctx {
		t.Fatal("Context was not set correctly")
	}
}

func TestServiceUrls(t *testing.T) {
	tests := []struct {
		name          string
		indexDataPath string
		wantDataFunc  func(baseUrl *url.URL) map[ServiceType]string
		want          bool
	}{
		{
			name:          "https://api.nuget.org/ index urls",
			indexDataPath: index_V3,
			wantDataFunc: func(baseUrl *url.URL) map[ServiceType]string {
				return map[ServiceType]string{
					SearchQueryService: fmt.Sprintf("%s://%s/query", baseUrl.Scheme, baseUrl.Host),
					RegistrationsBaseUrl: fmt.Sprintf(
						"%s://%s/v3/registration5-gz-semver2",
						baseUrl.Scheme,
						baseUrl.Host,
					),
					SearchAutocompleteService: fmt.Sprintf("%s://%s/autocomplete", baseUrl.Scheme, baseUrl.Host),
					ReportAbuseUriTemplate: fmt.Sprintf(
						"%s://%s/packages/{id}/{version}/ReportAbuse",
						baseUrl.Scheme,
						baseUrl.Host,
					),
					ReadmeUriTemplate: fmt.Sprintf(
						"%s://%s/v3-flatcontainer/{lower_id}/{lower_version}/readme",
						baseUrl.Scheme,
						baseUrl.Host,
					),
					PackageDetailsUriTemplate: fmt.Sprintf(
						"%s://%s/packages/{id}/{version}?_src=template",
						baseUrl.Scheme,
						baseUrl.Host,
					),
					LegacyGallery:      fmt.Sprintf("%s://%s/api/v2", baseUrl.Scheme, baseUrl.Host),
					PackagePublish:     fmt.Sprintf("%s://%s/api/v2/package", baseUrl.Scheme, baseUrl.Host),
					PackageBaseAddress: fmt.Sprintf("%s://%s/v3-flatcontainer", baseUrl.Scheme, baseUrl.Host),
					RepositorySignatures: fmt.Sprintf(
						"%s://%s/v3-index/repository-signatures/5.0.0/index.json",
						baseUrl.Scheme,
						baseUrl.Host,
					),
					SymbolPackagePublish: fmt.Sprintf(
						"%s://%s/api/v2/symbolpackage",
						baseUrl.Scheme,
						baseUrl.Host,
					),
					VulnerabilityInfo: fmt.Sprintf(
						"%s://%s/v3/vulnerabilities/index.json",
						baseUrl.Scheme,
						baseUrl.Host,
					),
					OwnerDetailsUriTemplate: fmt.Sprintf(
						"%s://%s/profiles/{owner}?_src=template",
						baseUrl.Scheme,
						baseUrl.Host,
					),
				}
			},
			want: true,
		},
		{
			name:          "baget index urls",
			indexDataPath: index_Baget,
			wantDataFunc: func(baseUrl *url.URL) map[ServiceType]string {
				return map[ServiceType]string{
					SearchQueryService:        fmt.Sprintf("%s://%s/v3/search", baseUrl.Scheme, baseUrl.Host),
					RegistrationsBaseUrl:      fmt.Sprintf("%s://%s/v3/registration", baseUrl.Scheme, baseUrl.Host),
					SearchAutocompleteService: fmt.Sprintf("%s://%s/v3/autocomplete", baseUrl.Scheme, baseUrl.Host),
					ReportAbuseUriTemplate: fmt.Sprintf(
						"%s://%s/packages/{id}/{version}/ReportAbuse",
						baseUrl.Scheme,
						baseUrl.Host,
					),
					ReadmeUriTemplate:         "",
					PackageDetailsUriTemplate: "",
					LegacyGallery:             "",
					PackagePublish:            fmt.Sprintf("%s://%s/api/v2/package", baseUrl.Scheme, baseUrl.Host),
					PackageBaseAddress:        fmt.Sprintf("%s://%s/v3/package", baseUrl.Scheme, baseUrl.Host),
					RepositorySignatures:      "",
					SymbolPackagePublish:      fmt.Sprintf("%s://%s/api/v2/symbol", baseUrl.Scheme, baseUrl.Host),
					VulnerabilityInfo:         "",
					OwnerDetailsUriTemplate:   "",
				}
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, client := setup(t, tc.indexDataPath)
			wantData := tc.wantDataFunc(client.BaseURL())

			urls := make(map[ServiceType]*url.URL)
			for st, item := range wantData {
				if item == "" {
					continue
				}
				u, err := url.Parse(item)
				require.NoError(t, err)
				urls[st] = u
			}
			if tc.want {
				require.Equal(t, urls, client.serviceUrls)
			} else {
				require.NotEqual(t, urls, client.serviceUrls)
			}
		})
	}
}
