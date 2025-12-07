package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// SendRequest sends a HTTP GET request to the specified URL.
// We use a different name than "Request" because "Request" is already a type.
func SendRequest(url string) Response {
	RecordAction(fmt.Sprintf("Request: %s", url), func() { SendRequest(url) })
	if IsDryRun() {
		return Response{}
	}
	Logf(LogTypeRequest, "Sending GET request to: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		// In integration tests, we often panic on setup failure, but here it's part of the test.
		// Returning a 0 status response might be confusing.
		// Let's panic to show the test failed immediately?
		// Or return error? The example: resp := Request(...); Expect...
		// If Request fails, Expect will likely fail or panic.
		Fail("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	header := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			header[k] = v[0]
		}
	}

	Log(LogTypeRequest, fmt.Sprintf("Received status %d from %s", resp.StatusCode, url), fmt.Sprintf("Body: %s\nHeaders: %v", string(body), header))
	return Response{
		StatusCode: resp.StatusCode,
		Body:       string(body),
		Header:     header,
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
