package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
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
