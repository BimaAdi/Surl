package core

import (
	"encoding/json"
	"fmt"

	"github.com/BimaAdi/surl/request"
)

func Substitute(value string, variables map[string]string) string {
	return variablePattern.ReplaceAllStringFunc(value, func(match string) string {
		name := match[1 : len(match)-1]
		if replacement, ok := variables[name]; ok {
			return replacement
		}
		return match
	})
}

func SubstituteRequest(spec RequestConfig, cfg Config) (RequestConfig, Config, error) {
	variables := cfg.Global.Variable

	cfg.Global.Headers = substituteMap(cfg.Global.Headers, variables)
	cfg.Global.BasicAuth = substituteBasicAuth(cfg.Global.BasicAuth, variables)

	spec.URL = Substitute(spec.URL, variables)
	spec.Query = substituteMap(spec.Query, variables)
	spec.Headers = substituteMap(spec.Headers, variables)
	spec.BasicAuth = substituteBasicAuth(spec.BasicAuth, variables)

	body, err := substituteRawBody(spec.Body, variables)
	if err != nil {
		return RequestConfig{}, Config{}, fmt.Errorf("substitute request body: %w", err)
	}
	spec.Body = body
	spec.JSON = json.RawMessage(Substitute(string(spec.JSON), variables))
	spec.Form = substituteFormValues(spec.Form, variables)
	spec.FormMultipart = substituteMultipartValues(spec.FormMultipart, variables)
	spec.MultipartForm = substituteMultipartValues(spec.MultipartForm, variables)

	return spec, cfg, nil
}

func substituteMap(values map[string]string, variables map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[Substitute(key, variables)] = Substitute(value, variables)
	}
	return result
}

func substituteBasicAuth(auth *BasicAuthConfig, variables map[string]string) *BasicAuthConfig {
	if auth == nil {
		return nil
	}
	return &BasicAuthConfig{
		Username: Substitute(auth.Username, variables),
		Password: Substitute(auth.Password, variables),
	}
}

// substituteRawBody keeps the body as a JSON string when it is one, so the
// raw body parser can still unwrap it after substitution.
func substituteRawBody(body json.RawMessage, variables map[string]string) (json.RawMessage, error) {
	var raw string
	if err := json.Unmarshal(body, &raw); err == nil {
		encoded, err := json.Marshal(Substitute(raw, variables))
		if err != nil {
			return nil, err
		}
		return encoded, nil
	}
	return json.RawMessage(Substitute(string(body), variables)), nil
}

func substituteFormValues(values []request.FormValue, variables map[string]string) []request.FormValue {
	if values == nil {
		return nil
	}
	result := make([]request.FormValue, len(values))
	for i, value := range values {
		result[i] = request.FormValue{
			Name:  Substitute(value.Name, variables),
			Value: Substitute(value.Value, variables),
		}
	}
	return result
}

func substituteMultipartValues(values []request.MultipartFormValue, variables map[string]string) []request.MultipartFormValue {
	if values == nil {
		return nil
	}
	result := make([]request.MultipartFormValue, len(values))
	for i, value := range values {
		result[i] = request.MultipartFormValue{
			Name:     Substitute(value.Name, variables),
			Type:     value.Type,
			Value:    Substitute(value.Value, variables),
			FileName: Substitute(value.FileName, variables),
			FilePath: Substitute(value.FilePath, variables),
		}
	}
	return result
}
