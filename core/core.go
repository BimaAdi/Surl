package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Config struct {
	Global struct {
		Variable map[string]string `json:"variable"`
	} `json:"global"`
	API map[string]RequestConfig `json:"api"`
}

type RequestConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Query   map[string]string `json:"query"`
	Body    json.RawMessage   `json:"body"`
	JSON    json.RawMessage   `json:"json"`
	Form    []FormValue       `json:"form"`
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

func ExecuteRequest(ctx context.Context, client HTTPClient, spec RequestConfig, variables map[string]string) (Response, error) {
	// Url
	requestURL := Substitute(spec.URL, variables)
	if requestURL == "" {
		return Response{}, errors.New("request URL is empty")
	}
	parsedURL, err := url.Parse(requestURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return Response{}, fmt.Errorf("invalid request URL %q", requestURL)
	}

	// Query parameters
	query := parsedURL.Query()
	for key, value := range spec.Query {
		query.Set(Substitute(key, variables), Substitute(value, variables))
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
	for key, value := range spec.Headers {
		req.Header.Set(Substitute(key, variables), Substitute(value, variables))
	}

	// ParsedBody
	parseRequestBody(req, spec, variables)

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

func parseRequestBody(req *http.Request, spec RequestConfig, variables map[string]string) {
	if len(spec.JSON) > 0 {
		parseJSONBody(req, spec.JSON, variables)
		return
	}
	if len(spec.Form) > 0 {
		parseFormBody(req, spec.Form, variables)
		return
	}
	parseRawBody(req, spec.Body, variables)
}

func Substitute(value string, variables map[string]string) string {
	return variablePattern.ReplaceAllStringFunc(value, func(match string) string {
		name := match[1 : len(match)-1]
		if replacement, ok := variables[name]; ok {
			return replacement
		}
		return match
	})
}
