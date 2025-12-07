package v1

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "Value")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"key": "value"}`)
	}))
	defer server.Close()

	resp := SendRequest(server.URL)

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if resp.Body != `{"key": "value"}` {
		t.Errorf("Expected body '{\"key\": \"value\"}', got '%s'", resp.Body)
	}
	if resp.Header["X-Test"] != "Value" {
		t.Errorf("Expected header X-Test=Value, got %s", resp.Header["X-Test"])
	}
}

func TestExpectFunctions(t *testing.T) {
	resp := Response{
		StatusCode: 200,
		Body:       `{"a": 1, "b": {"c": 2}, "d": [3, 4]}`,
		Header:     map[string]string{"Content-Type": "application/json"},
	}

	// Success cases (should not panic)
	ExpectStatusCode(resp, 200)
	ExpectHeader(resp, "Content-Type", "application/json")
	ExpectJsonBody(resp, `{"a": 1, "b": {"c": 2}, "d": [3, 4]}`)
	ExpectJsonBodyField(resp, "a", 1)
	ExpectJsonBodyField(resp, "b.c", 2)
	ExpectJsonBodyField(resp, "d[0]", 3)

	// Failure cases (should panic with TestError)
	assertPanic := func(name string, f func()) {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("%s expected to panic", name)
			}
			if _, ok := r.(TestError); !ok {
				t.Errorf("%s panicked with unexpected type: %T", name, r)
			}
		}()
		f()
	}

	assertPanic("ExpectStatusCode", func() { ExpectStatusCode(resp, 404) })
	assertPanic("ExpectHeader", func() { ExpectHeader(resp, "Content-Type", "xml") })
	assertPanic("ExpectJsonBody", func() { ExpectJsonBody(resp, `{"a": 2}`) })
	assertPanic("ExpectJsonBodyField", func() { ExpectJsonBodyField(resp, "a", 999) })
	assertPanic("ExpectJsonBodyField path", func() { ExpectJsonBodyField(resp, "x.y", 1) })
}
