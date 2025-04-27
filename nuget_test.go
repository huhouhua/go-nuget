// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
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
		f.Close()
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
	_, server := createHttpServer(t, "testdata/index.json")
	c, err := NewClient(WithBaseURL(server.URL))

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	for serviceType, url := range c.serviceUrls {
		t.Logf("type:%s url:%s", serviceType.String(), url.String())
	}
	//expectedBaseURL := defaultBaseURL+apiVersionPath

	if c.BaseURL().String() != server.URL {
		t.Errorf("NewClient BaseURL is %s, want %s", c.BaseURL().String(), server.URL)
	}
	if c.UserAgent != userAgent {
		t.Errorf("NewClient UserAgent is %s, want %s", c.UserAgent, userAgent)
	}
}

func TestCheckResponseOnHeadRequestError(t *testing.T) {
	_, server := createHttpServer(t, "testdata/index.json")
	c, err := NewClient(WithBaseURL(server.URL))

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	req, err := c.NewRequest(http.MethodHead, "test", nil, nil)
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
	_, server := createHttpServer(t, "testdata/index.json")
	c, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	req, err := c.NewRequest(http.MethodGet, "test", nil, []RequestOptionFunc{WithContext(ctx)})
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	defer cancel()

	if req.Context() != ctx {
		t.Fatal("Context was not set correctly")
	}
}

func TestServiceUrls(t *testing.T) {
	tests := []struct {
		name          string
		indexDataPath string
		wantData      map[ServiceType]string
		want          bool
	}{
		{
			name:          "https://api.nuget.org/ index urls",
			indexDataPath: "testdata/index.json",
			wantData: map[ServiceType]string{
				SearchQueryService:        "https://azuresearch-ussc.nuget.org/query",
				RegistrationsBaseUrl:      "https://api.nuget.org/v3/registration5-gz-semver2",
				SearchAutocompleteService: "https://azuresearch-ussc.nuget.org/autocomplete",
				ReportAbuseUriTemplate:    "https://www.nuget.org/packages/{id}/{version}/ReportAbuse",
				ReadmeUriTemplate:         "https://api.nuget.org/v3-flatcontainer/{lower_id}/{lower_version}/readme",
				PackageDetailsUriTemplate: "https://www.nuget.org/packages/{id}/{version}?_src=template",
				LegacyGallery:             "https://www.nuget.org/api/v2",
				PackagePublish:            "https://www.nuget.org/api/v2/package",
				PackageBaseAddress:        "https://api.nuget.org/v3-flatcontainer",
				RepositorySignatures:      "https://api.nuget.org/v3-index/repository-signatures/5.0.0/index.json",
				SymbolPackagePublish:      "https://www.nuget.org/api/v2/symbolpackage",
				VulnerabilityInfo:         "https://api.nuget.org/v3/vulnerabilities/index.json",
				OwnerDetailsUriTemplate:   "https://www.nuget.org/profiles/{owner}?_src=template",
			},
			want: true,
		},
		{
			name:          "baget index urls",
			indexDataPath: "testdata/index_2.json",
			wantData: map[ServiceType]string{
				SearchQueryService:        "http://localhost:5000/v3/search",
				RegistrationsBaseUrl:      "http://localhost:5000/v3/registration",
				SearchAutocompleteService: "http://localhost:5000/v3/autocomplete",
				ReportAbuseUriTemplate:    "https://www.nuget.org/packages/{id}/{version}/ReportAbuse",
				ReadmeUriTemplate:         "",
				PackageDetailsUriTemplate: "",
				LegacyGallery:             "",
				PackagePublish:            "http://localhost:5000/api/v2/package",
				PackageBaseAddress:        "http://localhost:5000/v3/package",
				RepositorySignatures:      "",
				SymbolPackagePublish:      "http://localhost:5000/api/v2/symbol",
				VulnerabilityInfo:         "",
				OwnerDetailsUriTemplate:   "",
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, client := setup(t, tc.indexDataPath)
			urls := make(map[ServiceType]*url.URL, len(tc.wantData))
			for st, item := range tc.wantData {
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
