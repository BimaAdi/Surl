package core

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/BimaAdi/surl/request"
)

type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TLSConfig struct {
	InsecureSkipVerify bool `json:"insecure-skip-verify"`
}

type Config struct {
	Global struct {
		Variable  map[string]string `json:"variable"`
		Headers   map[string]string `json:"headers"`
		BasicAuth *BasicAuthConfig  `json:"basic-auth"`
		TLSConfig *TLSConfig        `json:"tls-config"`
	} `json:"global"`
	API map[string]RequestConfig `json:"api"`
}

type RequestConfig struct {
	URL           string                       `json:"url"`
	Method        string                       `json:"method"`
	Headers       map[string]string            `json:"headers"`
	Query         map[string]string            `json:"query"`
	BasicAuth     *BasicAuthConfig             `json:"basic-auth"`
	TLSConfig     *TLSConfig                   `json:"tls-config"`
	Body          json.RawMessage              `json:"body"`
	JSON          json.RawMessage              `json:"json"`
	Form          []request.FormValue          `json:"form"`
	FormMultipart []request.MultipartFormValue `json:"form_multipart"`
	MultipartForm []request.MultipartFormValue `json:"multipart_form"`
}

type Response struct {
	Status int               `json:"status"`
	Header map[string]string `json:"header"`
	Body   string            `json:"body"`
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

var variablePattern = regexp.MustCompile(`\{([a-zA-Z0-9_.-]+)\}`)

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}
	return cfg, nil
}

func ExecuteRequest(ctx context.Context, client HTTPClient, spec RequestConfig, cfg Config) (Response, error) {
	globalHeaders := cfg.Global.Headers

	// Url
	parsedURL, err := url.Parse(spec.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return Response{}, fmt.Errorf("invalid request URL %q", spec.URL)
	}

	// Query parameters
	query := parsedURL.Query()
	for key, value := range spec.Query {
		query.Set(key, value)
	}
	parsedURL.RawQuery = query.Encode()

	// Method
	method := strings.ToUpper(spec.Method)
	if method == "" {
		method = http.MethodGet
	}

	// Init http module
	req, err := http.NewRequestWithContext(ctx, method, parsedURL.String(), nil)
	if err != nil {
		return Response{}, fmt.Errorf("create request: %w", err)
	}

	// Headers
	// Global headers are set first so local headers with the same name override them.
	for key, value := range globalHeaders {
		req.Header.Set(key, value)
	}
	for key, value := range spec.Headers {
		req.Header.Set(key, value)
	}

	// ParsedBody
	if err := parseRequestBody(req, spec); err != nil {
		return Response{}, fmt.Errorf("parse request body: %w", err)
	}

	// Basic auth
	// The API-level basic auth takes priority over the global one.
	basicAuth := spec.BasicAuth
	if basicAuth == nil {
		basicAuth = cfg.Global.BasicAuth
	}
	if basicAuth != nil {
		req.SetBasicAuth(basicAuth.Username, basicAuth.Password)
	}

	// TLS config
	// The API-level TLS config takes priority over the global one.
	tlsConfig := spec.TLSConfig
	if tlsConfig == nil {
		tlsConfig = cfg.Global.TLSConfig
	}
	if tlsConfig != nil {
		client = applyTLSConfig(client, tlsConfig)
	}

	// Execute and Parse the response
	httpResponse, err := client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("send request: %w", err)
	}
	defer httpResponse.Body.Close()
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return Response{}, fmt.Errorf("read response: %w", err)
	}
	headers := make(map[string]string, len(httpResponse.Header))
	for key, values := range httpResponse.Header {
		headers[key] = strings.Join(values, ", ")
	}
	return Response{Status: httpResponse.StatusCode, Header: headers, Body: string(responseBody)}, nil
}

func parseRequestBody(req *http.Request, spec RequestConfig) error {
	if len(spec.JSON) > 0 {
		request.ParseJSONBody(req, spec.JSON)
		return nil
	}
	if len(spec.Form) > 0 {
		request.ParseFormBody(req, spec.Form)
		return nil
	}
	multipartForm := spec.FormMultipart
	if len(multipartForm) == 0 {
		multipartForm = spec.MultipartForm
	}
	if len(multipartForm) > 0 {
		return request.ParseMultipartBody(req, multipartForm)
	}
	request.ParseRawBody(req, spec.Body)
	return nil
}

// applyTLSConfig returns a client whose transport uses the given TLS settings.
// When the client already uses a *http.Transport its settings are cloned and
// reused, otherwise the default transport settings are cloned.
func applyTLSConfig(client HTTPClient, tlsConfig *TLSConfig) HTTPClient {
	httpClient, ok := client.(*http.Client)
	if !ok {
		return client
	}
	base, ok := httpClient.Transport.(*http.Transport)
	if !ok {
		base, ok = http.DefaultTransport.(*http.Transport)
		if !ok {
			return client
		}
	}
	cloned := base.Clone()
	if cloned.TLSClientConfig == nil {
		cloned.TLSClientConfig = &tls.Config{}
	}
	cloned.TLSClientConfig.InsecureSkipVerify = tlsConfig.InsecureSkipVerify
	return &http.Client{
		Transport:     cloned,
		CheckRedirect: httpClient.CheckRedirect,
		Jar:           httpClient.Jar,
		Timeout:       httpClient.Timeout,
	}
}
