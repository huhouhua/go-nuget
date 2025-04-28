// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"math"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"golang.org/x/time/rate"
)

const (
	defaultBaseURL = "https://api.nuget.org/"
	apiVersionPath = "v3"
	userAgent      = "go-nuget"

	headerRateLimit             = "RateLimit-Limit"
	headerRateReset             = "RateLimit-Reset"
	DecoderTypeXML  DecoderType = "xml"
	DecoderTypeJSON DecoderType = "json"
	DecoderEmpty    DecoderType = ""
)

type DecoderType string

var ErrNotFound = errors.New("404 Not Found")

// A Client manages communication with the NuGet API.
// This package is completely frozen, nothing will be added, removed or changed.
type Client struct {
	// HTTP client used to communicate with the API.
	client *retryablehttp.Client

	// Base URL for API requests. Defaults to the public NuGet API, but can be
	// set to a domain endpoint to use with a self hosted NuGet server. baseURL
	// should always be specified with a trailing slash.
	baseURL *url.URL

	// disableRetries is used to disable the default retry logic.
	disableRetries bool

	// configureLimiterOnce is used to make sure the limiter is configured exactly
	// once and block all other calls until the initial (one) call is done.
	configureLimiterOnce sync.Once

	// Limiter is used to limit API calls and prevent 429 responses.
	limiter RateLimiter

	// apiKey used to make authenticated API calls.
	apiKey string

	// Protects the apiKey field from concurrent read/write accesses.
	apiKeyLock sync.RWMutex

	// serviceUrls is used to store the service Resource of the NuGet API.
	serviceUrls map[ServiceType]*url.URL

	// Default request options applied to every request.
	defaultRequestOptions []RequestOptionFunc

	// User agent used when communicating with the NuGet API.
	UserAgent string

	FindPackageResource *FindPackageResource

	MetadataResource *PackageMetadataResource

	SearchResource *PackageSearchResource

	UpdateResource *PackageUpdateResource

	IndexResource *ServiceResource
}

// RateLimiter describes the interface that all (custom) rate limiters must implement.
type RateLimiter interface {
	Wait(context.Context) error
}

func NewClient(options ...ClientOptionFunc) (*Client, error) {
	client, err := newClient(options...)
	if err != nil {
		return nil, err
	}
	return client, nil

}

// NewOAuthClient returns a new NuGet API client. To use API methods which
// require authentication, provide a valid oauth token.
// This package is completely frozen, nothing will be added, removed or changed.
func NewOAuthClient(apiKey string, options ...ClientOptionFunc) (*Client, error) {
	client, err := newClient(options...)
	if err != nil {
		return nil, err
	}
	client.apiKey = apiKey
	return client, nil
}

func newClient(options ...ClientOptionFunc) (*Client, error) {
	c := &Client{UserAgent: userAgent}

	// Configure the HTTP client.
	c.client = &retryablehttp.Client{
		Backoff:      c.retryHTTPBackoff,
		CheckRetry:   c.retryHTTPCheck,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
		HTTPClient:   cleanhttp.DefaultPooledClient(),
		RetryWaitMin: 100 * time.Millisecond,
		RetryWaitMax: 400 * time.Millisecond,
		RetryMax:     5,
	}

	// Set the default base URL.
	c.setBaseURL(defaultBaseURL)

	// Apply any given client options.
	for _, fn := range options {
		if fn == nil {
			continue
		}
		if err := fn(c); err != nil {
			return nil, err
		}
	}

	// If no custom limiter was set using a client option, configure
	// the default rate limiter with values that implicitly disable
	// rate limiting until an initial HTTP call is done and we can
	// use the headers to try and properly configure the limiter.
	if c.limiter == nil {
		c.limiter = rate.NewLimiter(rate.Inf, 0)
	}
	c.FindPackageResource = &FindPackageResource{client: c}
	c.MetadataResource = &PackageMetadataResource{client: c}
	c.SearchResource = &PackageSearchResource{client: c}
	c.UpdateResource = &PackageUpdateResource{client: c}
	c.IndexResource = &ServiceResource{client: c}

	c.serviceUrls = make(map[ServiceType]*url.URL)
	err := c.loadResource()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// BaseURL return a copy of the baseURL.
func (c *Client) BaseURL() *url.URL {
	u := *c.baseURL
	return &u
}

// retryHTTPCheck provides a callback for Client.CheckRetry which
// will retry both rate limit (429) and server (>= 500) errors.
func (c *Client) retryHTTPCheck(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		return false, err
	}
	if !c.disableRetries && (resp.StatusCode == 429 || resp.StatusCode >= 500) {
		return true, nil
	}
	return false, nil
}

// retryHTTPBackoff provides a generic callback for Client.Backoff which
// will pass through all calls based on the status code of the response.
func (c *Client) retryHTTPBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	// Use the rate limit backoff function when we are rate limited.
	if resp != nil && resp.StatusCode == 429 {
		return rateLimitBackoff(min, max, attemptNum, resp)
	}

	// Set custom duration's when we experience a service interruption.
	min = 700 * time.Millisecond
	max = 900 * time.Millisecond

	return retryablehttp.LinearJitterBackoff(min, max, attemptNum, resp)
}

// rateLimitBackoff provides a callback for Client.Backoff which will use the
// RateLimit-Reset header to determine the time to wait. We add some jitter
// to prevent a thundering herd.
//
// min and max are mainly used for bounding the jitter that will be added to
// the reset time retrieved from the headers. But if the final wait time is
// less then min, min will be used instead.
func rateLimitBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	// rnd is used to generate pseudo-random numbers.
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	// First create some jitter bounded by the min and max durations.
	jitter := time.Duration(rnd.Float64() * float64(max-min))

	if resp != nil {
		if v := resp.Header.Get(headerRateReset); v != "" {
			if reset, _ := strconv.ParseInt(v, 10, 64); reset > 0 {
				// Only update min if the given time to wait is longer.
				if wait := time.Until(time.Unix(reset, 0)); wait > min {
					min = wait
				}
			}
		} else {
			// In case the RateLimit-Reset header is not set, back off an additional
			// 100% exponentially. With the default milliseconds being set to 100 for
			// `min`, this makes the 5th retry wait 3.2 seconds (3,200 ms) by default.
			min = time.Duration(float64(min) * math.Pow(2, float64(attemptNum)))
		}
	}

	return min + jitter
}

// configureLimiter configures the rate limiter.
func (c *Client) configureLimiter(ctx context.Context, headers http.Header) {
	if v := headers.Get(headerRateLimit); v != "" {
		if rateLimit, _ := strconv.ParseFloat(v, 64); rateLimit > 0 {
			// The rate limit is based on requests per minute, so for our limiter to
			// work correctly we divide the limit by 60 to get the limit per second.
			rateLimit /= 60

			// Configure the limit and burst using a split of 2/3 for the limit and
			// 1/3 for the burst. This enables clients to burst 1/3 of the allowed
			// calls before the limiter kicks in. The remaining calls will then be
			// spread out evenly using intervals of time.Second / limit which should
			// prevent hitting the rate limit.
			limit := rate.Limit(rateLimit * 0.66)
			burst := int(rateLimit * 0.33)

			// Need at least one allowed to burst or x/time will throw an error
			if burst == 0 {
				burst = 1
			}

			// Create a new limiter using the calculated values.
			c.limiter = rate.NewLimiter(limit, burst)

			// Call the limiter once as we have already made a request
			// to get the headers and the limiter is not aware of this.
			c.limiter.Wait(ctx)
		}
	}
}

// setBaseURL sets the base URL for API requests to a custom endpoint.
func (c *Client) setBaseURL(urlStr string) error {
	// Make sure the given URL end with a slash
	//if !strings.HasSuffix(urlStr, "/") {
	//	urlStr += "/"
	//}

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	//if !strings.HasSuffix(baseURL.Path, apiVersionPath) {
	//	baseURL.Path += apiVersionPath
	//}

	// Update the base URL of the client.
	c.baseURL = baseURL

	return nil
}

// NewRequest creates a new API request. The method expects a relative URL
// path that will be resolved relative to the base URL of the Client.
// Relative URL paths should always be specified without a preceding slash.
// If specified, the value pointed to by body is JSON encoded and included
// as the request body.
func (c *Client) NewRequest(method, path string, baseUrl *url.URL, opt interface{}, options []RequestOptionFunc) (*retryablehttp.Request, error) {
	u := *c.baseURL
	if baseUrl != nil {
		u = *baseUrl
	}
	unescaped, err := url.PathUnescape(path)
	if err != nil {
		return nil, err
	}

	// Set the encoded path data
	u.RawPath = c.baseURL.Path + path
	u.Path = pathCombine(c.baseURL.Path, unescaped)

	// Create a request specific headers map.
	reqHeaders := make(http.Header)
	reqHeaders.Set("Accept", "application/json")

	if c.UserAgent != "" {
		reqHeaders.Set("User-Agent", c.UserAgent)
	}

	var body interface{}
	switch {
	case method == http.MethodPatch || method == http.MethodPost || method == http.MethodPut:
		reqHeaders.Set("Content-Type", "application/json")

		if opt != nil {
			body, err = json.Marshal(opt)
			if err != nil {
				return nil, err
			}
		}
	case opt != nil:
		q, err := query.Values(opt)
		if err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
	}

	req, err := retryablehttp.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	for _, fn := range append(c.defaultRequestOptions, options...) {
		if fn == nil {
			continue
		}
		if err := fn(req); err != nil {
			return nil, err
		}
	}

	// Set the request specific headers.
	for k, v := range reqHeaders {
		req.Header[k] = v
	}

	return req, nil
}

// UploadRequest creates an API request for uploading a file. The method
// expects a relative URL path that will be resolved relative to the base
// URL of the Client. Relative URL paths should always be specified without
// a preceding slash. If specified, the value pointed to by body is JSON
// encoded and included as the request body.
func (c *Client) UploadRequest(method, path string, baseUrl *url.URL, content io.Reader, fileType, filename string, opt interface{}, options []RequestOptionFunc) (*retryablehttp.Request, error) {
	u := *c.baseURL
	if baseUrl != nil {
		u = *baseUrl
	}
	unescaped, err := url.PathUnescape(path)
	if err != nil {
		return nil, err
	}

	// Set the encoded path data
	u.RawPath = c.baseURL.Path + path
	u.Path = pathCombine(c.baseURL.Path, unescaped)

	// Create a request specific headers map.
	reqHeaders := make(http.Header)
	reqHeaders.Set("Accept", "application/json")
	reqHeaders.Set("Transfer-Encoding", "chunked")
	if c.UserAgent != "" {
		reqHeaders.Set("User-Agent", c.UserAgent)
	}

	b := new(bytes.Buffer)
	w := multipart.NewWriter(b)

	fw, err := w.CreateFormFile(fileType, filename)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(fw, content); err != nil {
		return nil, err
	}

	if opt != nil {
		fields, err := query.Values(opt)
		if err != nil {
			return nil, err
		}
		for name := range fields {
			if err = w.WriteField(name, fmt.Sprintf("%v", fields.Get(name))); err != nil {
				return nil, err
			}
		}
	}

	if err = w.Close(); err != nil {
		return nil, err
	}

	reqHeaders.Set("Content-Type", w.FormDataContentType())

	req, err := retryablehttp.NewRequest(method, u.String(), b)
	if err != nil {
		return nil, err
	}

	for _, fn := range append(c.defaultRequestOptions, options...) {
		if fn == nil {
			continue
		}
		if err := fn(req); err != nil {
			return nil, err
		}
	}

	// Set the request specific headers.
	for k, v := range reqHeaders {
		req.Header[k] = v
	}

	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (c *Client) Do(req *retryablehttp.Request, v interface{}, decoder DecoderType) (*http.Response, error) {
	// Wait will block until the limiter can obtain a new token.
	err := c.limiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}

	// Set the correct authentication header. If using basic auth, then check
	// if we already have a token and if not first authenticate and get one.
	if values := req.Header.Values("X-NuGet-ApiKey"); len(values) == 0 {
		req.Header.Set("X-NuGet-ApiKey", c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	// If not yet configured, try to configure the rate limiter
	// using the response headers we just received. Fail silently
	// so the limiter will remain disabled in case of an error.
	c.configureLimiterOnce.Do(func() { c.configureLimiter(req.Context(), resp.Header) })

	err = CheckResponse(resp)
	if err != nil {
		// Even though there was an error, we still return the response
		// in case the caller wants to inspect it further.
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = decoder.Invoke(resp.Body, v)
		}
	}

	return resp, err
}

// loadResource loads the service index resource.
func (c *Client) loadResource() error {
	if c.IndexResource == nil {
		return fmt.Errorf("IndexResource is null")
	}
	index, _, err := c.IndexResource.GetIndex()
	if err != nil {
		return err
	}
	resourceMap := make(map[string]string)
	for _, resource := range index.Resources {
		if resource.Type != "" && resource.Id != "" {
			resourceMap[resource.Type] = resource.Id
		}
	}
	for t, typeVariants := range typesMap {
		var find bool
		for _, variant := range typeVariants.Types {
			if id, ok := resourceMap[variant]; ok {
				u, err := url.Parse(strings.TrimSuffix(id, "/"))
				if err != nil {
					return err
				}
				c.serviceUrls[t] = u
				find = true
				break
			}
		}
		if !find && typeVariants.DefaultUrl != "" {
			u, err := url.Parse(strings.TrimSuffix(typeVariants.DefaultUrl, "/"))
			if err != nil {
				return err
			}
			c.serviceUrls[t] = u
		}
	}
	return nil
}

// getResourceUrl returns the resource URL for the given resource value.
func (c *Client) getResourceUrl(value ServiceType) *url.URL {
	if u, ok := c.serviceUrls[value]; ok {
		return u
	}
	return nil
}

func (d DecoderType) Invoke(r io.Reader, v interface{}) error {
	switch d {
	case DecoderTypeXML:
		return xml.NewDecoder(r).Decode(v)
	case DecoderTypeJSON:
		return json.NewDecoder(r).Decode(v)
	default:
		return fmt.Errorf("unsupported decoder type: %s", d)
	}
}
func parseID(id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("id is null or empty")
	}
	return strings.ToLower(id), nil
}

// PathEscape Helper function to escape a packageId identifier.
func PathEscape(s string) string {
	return strings.ReplaceAll(url.PathEscape(s), ".", "%2E")
}

// An ErrorResponse reports one or more errors caused by an API request.
type ErrorResponse struct {
	Body     []byte
	Response *http.Response
	Message  string
}

func (e *ErrorResponse) Error() string {
	path, _ := url.QueryUnescape(e.Response.Request.URL.Path)
	url := fmt.Sprintf("%s://%s%s", e.Response.Request.URL.Scheme, e.Response.Request.URL.Host, path)

	if e.Message == "" {
		return fmt.Sprintf("%s %s: %d", e.Response.Request.Method, url, e.Response.StatusCode)
	} else {
		return fmt.Sprintf("%s %s: %d %s", e.Response.Request.Method, url, e.Response.StatusCode, e.Message)
	}
}

// CheckResponse checks the API response for errors, and returns them if present.
func CheckResponse(r *http.Response) error {
	switch r.StatusCode {
	case 200, 201, 202, 204, 304:
		return nil
	case 404:
		return ErrNotFound
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil && strings.TrimSpace(string(data)) != "" {
		errorResponse.Body = data
		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			errorResponse.Message = fmt.Sprintf("failed to parse unknown error format: %s %s", data, r.Status)
		} else {
			errorResponse.Message = parseError(raw)
		}
	}

	return errorResponse
}

func parseError(raw interface{}) string {
	switch raw := raw.(type) {
	case string:
		return raw

	case []interface{}:
		var errs []string
		for _, v := range raw {
			errs = append(errs, parseError(v))
		}
		return fmt.Sprintf("[%s]", strings.Join(errs, ", "))

	case map[string]interface{}:
		var errs []string
		for k, v := range raw {
			errs = append(errs, fmt.Sprintf("{%s: %s}", k, parseError(v)))
		}
		sort.Strings(errs)
		return strings.Join(errs, ", ")

	default:
		return fmt.Sprintf("failed to parse unexpected error type: %T", raw)
	}
}
