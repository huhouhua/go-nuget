// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
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
		WithSourceURL(fmt.Sprintf("%s/v3/index.json", server.URL)),
		// Disable backoff to speed up tests that expect errors.
		WithBackoff(func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
			return 0
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	for _, u := range client.serviceURLs {
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

func createFile(t *testing.T, path, data string) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(t, err)

	err = os.WriteFile(path, []byte(data), 0644)
	require.NoError(t, err)
}

func TestNewClient(t *testing.T) {
	_, server := createHttpServer(t, index_V3)

	sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
	c, err := NewClient(WithSourceURL(sourceURL))

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	for serviceType, u := range c.serviceURLs {
		t.Logf("type:%s url:%s", serviceType.String(), u.String())
	}
	//expectedBaseURL := defaultBaseURL+apiVersionPath

	if c.SourceURL().String() != sourceURL {
		t.Errorf("NewClient BaseURL is %s, want %s", c.SourceURL().String(), sourceURL)
	}
	if c.UserAgent != userAgent {
		t.Errorf("NewClient UserAgent is %s, want %s", c.UserAgent, userAgent)
	}

	err = c.setSourceURL("http://abc/%eth")
	wantErr := &url.Error{
		Op:  "parse",
		URL: "http://abc/%eth",
		Err: url.EscapeError("%et"),
	}
	require.Equalf(t, wantErr, err, "NewClient BaseURL is %+v, want %+v", err, wantErr)
}

func TestNewOAuthClient(t *testing.T) {
	t.Run("new a oath client return success", func(t *testing.T) {
		_, server := createHttpServer(t, index_V3)

		sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
		c, err := NewOAuthClient("", WithSourceURL(sourceURL))

		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		if c.SourceURL().String() != sourceURL {
			t.Errorf("NewOAuthClient BaseURL is %s, want %s", c.SourceURL().String(), sourceURL)
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

	sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
	c, err := NewClient(WithSourceURL(sourceURL))
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

	sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
	c, err := NewClient(WithSourceURL(sourceURL))
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

func TestSendRequest(t *testing.T) {
	mux, server := createHttpServer(t, index_V3)

	sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
	c, err := NewClient(WithSourceURL(sourceURL))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	req, err := c.NewRequest(http.MethodHead, "test", nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, `{"Name":"test"}`)
		require.NoError(t, err)
	})
	t.Run("limiter wait return error", func(t *testing.T) {
		c.limiter = new(errorRateLimiter)
		_, err = c.Do(req, new(interface{}), DecoderEmpty)
		wantErr := fmt.Errorf("wait fail")
		require.Equal(t, wantErr, err)
	})
	t.Run("unsupported decoder type return error", func(t *testing.T) {
		c.limiter = rate.NewLimiter(rate.Inf, 0)
		vMap := map[string]string{}
		_, err = c.Do(req, &vMap, "unsupported")
		wantErr := fmt.Errorf("unsupported decoder type: unsupported")
		require.Equal(t, wantErr, err)
	})
}

func TestCheckResponseOnHeadRequestError(t *testing.T) {
	_, server := createHttpServer(t, index_V3)

	sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
	c, err := NewClient(WithSourceURL(sourceURL))
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

func TestCheckResponseOnUnknownErrorFormat(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	req, err := c.NewRequest(http.MethodGet, "test", nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp := &http.Response{
		Request:    req.Request,
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("some error message but not JSON")),
	}

	errResp := CheckResponse(resp)
	if errResp == nil {
		t.Fatal("Expected error response.")
	}

	want := "GET https://api.nuget.org/test: 400 failed to parse unknown error format: some error message but not JSON "

	if errResp.Error() != want {
		t.Errorf("Expected error: %s, got %s", want, errResp.Error())
	}
	errResp = &ErrorResponse{
		Message:  "",
		Response: resp,
	}
	want = "GET https://api.nuget.org/test: 400"
	require.Equal(t, want, errResp.Error())
}

func TestParseError(t *testing.T) {
	t.Run("parse return success", func(t *testing.T) {
		input := []interface{}{
			"test",
		}
		expected := parseError(input)
		actual := "[test]"
		require.Equal(t, actual, expected)
	})
	t.Run("unexpected type return error", func(t *testing.T) {
		input := new(interface{})
		expected := parseError(input)
		actual := "failed to parse unexpected error type: *interface {}"
		require.Equal(t, actual, expected)
	})
}

func TestNewRequest(t *testing.T) {
	_, server := createHttpServer(t, index_V3)

	sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
	c, err := NewClient(WithSourceURL(sourceURL))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	t.Run("marshal return error", func(t *testing.T) {
		opt := struct {
			Ch chan int
		}{
			Ch: make(chan int),
		}
		_, actualErr := c.NewRequest(http.MethodPost, "test", nil, &opt, nil)
		require.IsType(t, new(json.UnsupportedTypeError), actualErr)
	})
	t.Run("values return error", func(t *testing.T) {
		_, actualErr := c.NewRequest(http.MethodGet, "test", nil, new(interface{}), nil)
		wantErr := errors.New("query: Values() expects struct input. Got interface")
		require.Equal(t, wantErr, actualErr)
	})
	t.Run("new request return error", func(t *testing.T) {
		u := createUrl(t, "http://localhost:5000")
		u.Scheme = "://abc"
		_, actualErr := c.NewRequest(http.MethodGet, "test", u, nil, nil)
		wantErr := &url.Error{
			Op:  "parse",
			URL: u.String() + "/test",
			Err: errors.New("missing protocol scheme"),
		}
		require.Equal(t, wantErr, actualErr)
	})
}

func TestUploadRequest(t *testing.T) {
	_, server := createHttpServer(t, index_V3)

	sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
	c, err := NewClient(WithSourceURL(sourceURL))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	t.Run("writer content return error", func(t *testing.T) {
		_, actualErr := c.UploadRequest(http.MethodGet, "", nil,
			os.Stderr, "", "", nil, nil)
		wantErr := &os.PathError{
			Op:   "read",
			Path: "/dev/stderr",
			Err:  syscall.Errno(9),
		}
		require.Equal(t, wantErr, actualErr)
	})
	t.Run("values return error", func(t *testing.T) {
		_, actualErr := c.UploadRequest(http.MethodGet, "test", nil,
			strings.NewReader("test"), "", "", new(interface{}), nil)
		wantErr := errors.New("query: Values() expects struct input. Got interface")
		require.Equal(t, wantErr, actualErr)
	})
	t.Run("new request return error", func(t *testing.T) {
		u := createUrl(t, "http://localhost:5000")
		u.Scheme = "://abc"
		_, actualErr := c.UploadRequest(http.MethodGet, "test", u,
			strings.NewReader("test"), "", "", nil, nil)
		wantErr := &url.Error{
			Op:  "parse",
			URL: u.String() + "/test",
			Err: errors.New("missing protocol scheme"),
		}
		require.Equal(t, wantErr, actualErr)
	})
	t.Run("request options return error", func(t *testing.T) {
		u := createUrl(t, "http://localhost:5000")
		wantErr := errors.New("request options fail")
		_, actualErr := c.UploadRequest(
			http.MethodGet,
			"test",
			u,
			strings.NewReader(
				"test",
			),
			"",
			"",
			nil,
			[]RequestOptionFunc{nil, func(request *retryablehttp.Request) error {
				return wantErr
			}},
		)
		require.Equal(t, wantErr, actualErr)
	})

	t.Run("write field return success", func(t *testing.T) {
		u := createUrl(t, "http://localhost:5000")
		opt := struct {
			Name string
		}{
			Name: "test",
		}

		r, err := c.UploadRequest(
			http.MethodGet,
			"test",
			u,
			strings.NewReader(
				"test",
			),
			"",
			"",
			&opt,
			nil,
		)
		require.NoError(t, err)
		require.NotNil(t, r)
	})
}

func TestRequestWithContext(t *testing.T) {
	_, server := createHttpServer(t, index_V3)

	sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
	c, err := NewClient(WithSourceURL(sourceURL))
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
func TestPathEscape(t *testing.T) {
	want := "diaspora%2Fdiaspora"
	got := PathEscape("diaspora/diaspora")
	if want != got {
		t.Errorf("Expected: %s, got %s", want, got)
	}
}

func TestLoadResource(t *testing.T) {
	tests := []struct {
		name             string
		configClientFunc func(client *Client)
		handleFunc       func(http.ResponseWriter, *http.Request)
		wantErr          error
	}{
		{
			name: "indexResource is null return error",
			configClientFunc: func(client *Client) {
				client.IndexResource = nil
			},
			wantErr: fmt.Errorf("IndexResource is null"),
		},
		{
			name: "url parse return error",
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				svc := &ServiceIndex{
					Resources: []*Resource{
						{
							Id:   "http://localhost:5000/query/%eth",
							Type: "SearchQueryService/Versioned",
						},
					},
				}
				data, err := json.Marshal(svc)
				require.NoError(t, err)
				_, err = w.Write(data)
				require.NoError(t, err)
			},
			wantErr: &url.Error{
				Op:  "parse",
				URL: "http://localhost:5000/query/%eth",
				Err: url.EscapeError("%et"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			if tc.handleFunc != nil {
				mux.HandleFunc("/v3/index.json", tc.handleFunc)
			}
			// server is a test HTTP server used to provide mock API responses.
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			c := &Client{
				client: retryablehttp.NewClient(),
			}
			c.IndexResource = &ServiceResource{client: c}
			c.limiter = rate.NewLimiter(rate.Inf, 0)

			sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
			err := c.setSourceURL(sourceURL)

			require.NoError(t, err)
			c.serviceURLs = make(map[ServiceType]*url.URL)

			if tc.configClientFunc != nil {
				tc.configClientFunc(c)
			}
			err = c.loadResource()
			require.Equal(t, tc.wantErr, err)
		})
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
					RegistrationsBaseURL: fmt.Sprintf(
						"%s://%s/v3/registration5-gz-semver2",
						baseUrl.Scheme,
						baseUrl.Host,
					),
					SearchAutocompleteService: fmt.Sprintf("%s://%s/autocomplete", baseUrl.Scheme, baseUrl.Host),
					ReportAbuseURLTemplate: fmt.Sprintf(
						"%s://%s/packages/{id}/{version}/ReportAbuse",
						baseUrl.Scheme,
						baseUrl.Host,
					),
					ReadmeURLTemplate: fmt.Sprintf(
						"%s://%s/v3-flatcontainer/{lower_id}/{lower_version}/readme",
						baseUrl.Scheme,
						baseUrl.Host,
					),
					PackageDetailsURLTemplate: fmt.Sprintf(
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
					OwnerDetailsURLTemplate: fmt.Sprintf(
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
			wantDataFunc: func(sourceURL *url.URL) map[ServiceType]string {
				return map[ServiceType]string{
					SearchQueryService:        fmt.Sprintf("%s://%s/v3/search", sourceURL.Scheme, sourceURL.Host),
					RegistrationsBaseURL:      fmt.Sprintf("%s://%s/v3/registration", sourceURL.Scheme, sourceURL.Host),
					SearchAutocompleteService: fmt.Sprintf("%s://%s/v3/autocomplete", sourceURL.Scheme, sourceURL.Host),
					ReportAbuseURLTemplate: fmt.Sprintf(
						"%s://%s/packages/{id}/{version}/ReportAbuse",
						sourceURL.Scheme,
						sourceURL.Host,
					),
					ReadmeURLTemplate:         "",
					PackageDetailsURLTemplate: "",
					LegacyGallery:             "",
					PackagePublish:            fmt.Sprintf("%s://%s/api/v2/package", sourceURL.Scheme, sourceURL.Host),
					PackageBaseAddress:        fmt.Sprintf("%s://%s/v3/package", sourceURL.Scheme, sourceURL.Host),
					RepositorySignatures:      "",
					SymbolPackagePublish:      fmt.Sprintf("%s://%s/api/v2/symbol", sourceURL.Scheme, sourceURL.Host),
					VulnerabilityInfo:         "",
					OwnerDetailsURLTemplate:   "",
				}
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, client := setup(t, tc.indexDataPath)
			wantData := tc.wantDataFunc(client.SourceURL())

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
				require.Equal(t, urls, client.serviceURLs)
			} else {
				require.NotEqual(t, urls, client.serviceURLs)
			}
		})
	}
}

type errorRateLimiter struct {
}

func (t *errorRateLimiter) Wait(context.Context) error {
	return fmt.Errorf("wait fail")
}
