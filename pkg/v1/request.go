package v1

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// SendRESTRequest sends an HTTP request with flexible options.
// Common usage:
//
//	SendRESTRequest(url,
//	    WithMethod("POST"),
//	    WithHeader("Authorization", "Bearer ..."),
//	    WithJSONBody(map[string]interface{}{...}),
//	    WithIgnoreServerSSL(true),
//	)
func SendRESTRequest(url string, opts ...RESTRequestOption) Response {
	cfg := restRequestConfig{
		method:  http.MethodGet,
		headers: make(map[string]string),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	RecordAction(fmt.Sprintf("Request: %s %s", cfg.method, url), func() {
		SendRESTRequest(url, opts...)
	})
	if IsDryRun() {
		return Response{}
	}

	var bodyReader io.Reader
	if len(cfg.body) > 0 {
		bodyReader = bytes.NewReader(cfg.body)
	}

	req, err := http.NewRequest(cfg.method, url, bodyReader)
	if err != nil {
		Fail("Request build failed: %v", err)
	}

	for k, v := range cfg.headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	if cfg.ignoreServerSSL {
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}

	requestBody := string(cfg.body)
	requestPrettyBody := requestBody
	if len(cfg.body) > 0 {
		var jsonObj interface{}
		if json.Unmarshal(cfg.body, &jsonObj) == nil {
			if pretty, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
				requestPrettyBody = string(pretty)
			}
		}
	}

	reqHeaderLines := make([]string, 0, len(cfg.headers))
	for k, v := range cfg.headers {
		if v == "" {
			reqHeaderLines = append(reqHeaderLines, fmt.Sprintf("%s:", k))
			continue
		}
		reqHeaderLines = append(reqHeaderLines, fmt.Sprintf("%s: %s", k, v))
	}

	Log(LogTypeRequest, fmt.Sprintf("Sending %s request to: %s", cfg.method, url), fmt.Sprintf("Body:\n%s\nHeaders:\n%s", requestPrettyBody, strings.Join(reqHeaderLines, "\n")))
	resp, err := client.Do(req)
	if err != nil {
		Fail("Request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	prettyBody := string(respBody)
	if len(respBody) > 0 {
		var jsonObj interface{}
		if json.Unmarshal(respBody, &jsonObj) == nil {
			if pretty, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
				prettyBody = string(pretty)
			}
		}
	}

	header := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			header[k] = v[0]
		}
	}

	headerLines := make([]string, 0, len(resp.Header))
	for k, v := range resp.Header {
		if len(v) == 0 {
			headerLines = append(headerLines, fmt.Sprintf("%s:", k))
			continue
		}
		for _, vv := range v {
			headerLines = append(headerLines, fmt.Sprintf("%s: %s", k, vv))
		}
	}

	Log(LogTypeRequest, fmt.Sprintf("Received status %d from %s", resp.StatusCode, url), fmt.Sprintf("Body:\n%s\nHeaders:\n%s", prettyBody, strings.Join(headerLines, "\n")))
	return Response{
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
		Header:     header,
	}
}

// SendRequest keeps backward compatibility; it is equivalent to GET via SendRESTRequest.
func SendRequest(url string) Response {
	return SendRESTRequest(url)
}

// RESTRequestOption configures SendRESTRequest.
type RESTRequestOption func(*restRequestConfig)

type restRequestConfig struct {
	method          string
	headers         map[string]string
	body            []byte
	ignoreServerSSL bool
}

// WithMethod sets HTTP method (GET by default).
func WithMethod(method string) RESTRequestOption {
	return func(c *restRequestConfig) {
		if method != "" {
			c.method = method
		}
	}
}

// WithHeader adds a single header.
func WithHeader(key, value string) RESTRequestOption {
	return func(c *restRequestConfig) {
		if c.headers == nil {
			c.headers = make(map[string]string)
		}
		c.headers[key] = value
	}
}

// WithHeaders merges multiple headers.
func WithHeaders(headers map[string]string) RESTRequestOption {
	return func(c *restRequestConfig) {
		if c.headers == nil {
			c.headers = make(map[string]string)
		}
		for k, v := range headers {
			c.headers[k] = v
		}
	}
}

// WithJSONBody marshals the given value as JSON and sets it as body.
// It also sets Content-Type to application/json if not already provided.
func WithJSONBody(v interface{}) RESTRequestOption {
	return func(c *restRequestConfig) {
		if v == nil {
			return
		}
		data, err := json.Marshal(v)
		if err != nil {
			Fail("Failed to marshal JSON body: %v", err)
		}
		c.body = data
		if _, ok := c.headers["Content-Type"]; !ok {
			c.headers["Content-Type"] = "application/json"
		}
	}
}

// WithBody sets raw bytes as body (caller can set headers accordingly).
func WithBody(body []byte) RESTRequestOption {
	return func(c *restRequestConfig) {
		c.body = body
	}
}

// WithBodyString sets body from string.
func WithBodyString(body string) RESTRequestOption {
	return func(c *restRequestConfig) {
		c.body = []byte(body)
	}
}

// WithIgnoreServerSSL skips server certificate verification (useful for tests/self-signed certs).
func WithIgnoreServerSSL(ignore bool) RESTRequestOption {
	return func(c *restRequestConfig) {
		c.ignoreServerSSL = ignore
	}
}

// ExpectStatusCode asserts that the response status code matches the expected code.
func ExpectStatusCode(resp Response, expected int) {
	if IsDryRun() {
		return
	}
	if resp.StatusCode != expected {
		// Include body in failure message for debugging
		Fail("Expected Status Code %d, got %d. Body: %s", expected, resp.StatusCode, resp.Body)
	}
	Logf(LogTypeExpect, "Status Code %d == %d - PASSED", resp.StatusCode, expected)
}

// ExpectHeader asserts that the response has the expected header.
func ExpectHeader(resp Response, key, value string) {
	if IsDryRun() {
		return
	}
	if got, ok := resp.Header[key]; !ok || got != value {
		Fail("ExpectHeader failed: expected %s=%s, got %s", key, value, got)
	}
	Logf(LogTypeExpect, "Header '%s' == '%s' - PASSED", key, value)
}

// ExpectJsonBody asserts that the response body matches the expected JSON.
// This is a simple implementation that compares unmarshaled objects.
func ExpectJsonBody(resp Response, expectedJson interface{}) {
	if IsDryRun() {
		return
	}
	var got interface{}
	if err := json.Unmarshal([]byte(resp.Body), &got); err != nil {
		Fail("ExpectJsonBody failed: response body is not valid JSON: %v. Body: %s", err, resp.Body)
	}

	// If expectedJson is string, unmarshal it too
	var expected interface{}
	if s, ok := expectedJson.(string); ok {
		if err := json.Unmarshal([]byte(s), &expected); err != nil {
			Fail("ExpectJsonBody failed: expectedJson string is not valid JSON: %v", err)
		}
	} else {
		expected = expectedJson
	}

	if !reflect.DeepEqual(got, expected) {
		Fail("ExpectJsonBody failed:\nExpected: %v\nGot:      %v", expected, got)
	}
	Log(LogTypeExpect, "JSON body matches expected value - PASSED", "")
}

// ExpectJsonBodyField asserts that a specific field in the JSON response body matches the expected value.
// field supports dot notation and array index (e.g. "data.users[0].name")
func ExpectJsonBodyField(resp Response, field string, expectedValue interface{}) {
	if IsDryRun() {
		return
	}

	var body interface{}
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		Fail("ExpectJsonBodyField failed: response body is not valid JSON: %v. Body: %s", err, resp.Body)
	}

	gotValue, err := getValueByPath(body, field)
	if err != nil {
		Fail("ExpectJsonBodyField failed to get field '%s': %v. Body: %s", field, err, resp.Body)
	}

	match := false
	if isNumber(gotValue) && isNumber(expectedValue) {
		if toFloat64(gotValue) == toFloat64(expectedValue) {
			match = true
		}
	} else {
		if reflect.DeepEqual(gotValue, expectedValue) {
			match = true
		}
	}

	if !match {
		Fail("ExpectJsonBodyField failed for field '%s':\nExpected: %v (%T)\nGot:      %v (%T)", field, expectedValue, expectedValue, gotValue, gotValue)
	}
	Logf(LogTypeExpect, "JSON Field '%s' == %v - PASSED", field, expectedValue)
}

func getValueByPath(data interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		key := part
		index := -1

		if strings.Contains(part, "[") && strings.HasSuffix(part, "]") {
			open := strings.Index(part, "[")
			close := strings.LastIndex(part, "]")
			if open > 0 {
				key = part[:open]
			} else {
				key = "" // e.g. [0] at start or after dot
			}
			idxStr := part[open+1 : close]
			var err error
			index, err = strconv.Atoi(idxStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index in path segment '%s': %v", part, err)
			}
		}

		if key != "" {
			m, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected map at '%s' but got %T (value: %v)", key, current, current)
			}
			val, exists := m[key]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found", key)
			}
			current = val
		}

		if index >= 0 {
			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array for index [%d] at '%s' but got %T (value: %v)", index, part, current, current)
			}
			if index >= len(arr) {
				return nil, fmt.Errorf("index %d out of bounds (len: %d)", index, len(arr))
			}
			current = arr[index]
		}
	}
	return current, nil
}

func isNumber(i interface{}) bool {
	switch i.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	}
	return false
}

func toFloat64(i interface{}) float64 {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		return v.Float()
	}
	return 0
}
