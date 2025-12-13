package v1

import (
	"fmt"
	"io"
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

func TestExpectJsonBodyFieldCond(t *testing.T) {
	resp := Response{
		Body: `{"num": 5, "text": "hello world", "nullField": null, "nested": {"arr": [1, 2]}}`,
	}

	// Success cases
	ExpectJsonBodyFieldCond(resp, "num", ConditionGreaterThan, 3)
	ExpectJsonBodyFieldCond(resp, "num", ConditionLessThanOrEqual, 5)
	ExpectJsonBodyFieldCond(resp, "text", ConditionContains, "hello")
	ExpectJsonBodyFieldCond(resp, "text", ConditionStartsWith, "hello")
	ExpectJsonBodyFieldCond(resp, "text", ConditionEndsWith, "world")
	ExpectJsonBodyFieldCond(resp, "nested.arr[1]", ConditionEqual, 2)
	ExpectJsonBodyFieldCond(resp, "nullField", ConditionEqual, nil)
	ExpectJsonBodyFieldCond(resp, "nullField", ConditionNotEqual, "not-nil")

	// Failure cases (should panic)
	assertPanic := func(name string, f func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("%s expected to panic", name)
			}
		}()
		f()
	}

	assertPanic("invalid path", func() { ExpectJsonBodyFieldCond(resp, "missing", ConditionEqual, 1) })
	assertPanic("condition mismatch", func() { ExpectJsonBodyFieldCond(resp, "num", ConditionLessThan, 1) })
}

func TestSendRESTRequestWithMethodHeadersAndJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}
		if r.Header.Get("X-Test") != "Value" {
			t.Errorf("expected header X-Test=Value, got %s", r.Header.Get("X-Test"))
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"a":1}` {
			t.Errorf("expected body {\"a\":1}, got %s", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	resp := SendRESTRequest(server.URL,
		WithMethod(http.MethodPost),
		WithHeader("X-Test", "Value"),
		WithJSONBody(map[string]int{"a": 1}),
	)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	if resp.Body != `{"ok":true}` {
		t.Fatalf("expected body '{\"ok\":true}', got %s", resp.Body)
	}
	if resp.Header["Content-Type"] != "application/json" {
		t.Fatalf("expected content-type application/json, got %s", resp.Header["Content-Type"])
	}
}

func TestSendRESTRequestIgnoreSSL(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "secure")
	}))
	defer server.Close()

	resp := SendRESTRequest(server.URL, WithIgnoreServerSSL(true))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if resp.Body != "secure" {
		t.Fatalf("expected body 'secure', got %s", resp.Body)
	}
}
