package dynamic_mock_server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerExecutor_PrepareData(t *testing.T) {
	t.Run("IfRequestHeader", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Auth", "Bearer 123")
		w := httptest.NewRecorder()
		h := NewHandlerExecutor(w, req)

		steps := []ResponseFuncConfig{
			IfRequestHeader("Auth", "Equal", "Bearer 123", "AUTH_OK", "true"),
		}
		if err := h.Execute(steps); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		if val, ok := h.Variables["AUTH_OK"]; !ok || val != "true" {
			t.Errorf("Expected AUTH_OK=true, got %v", val)
		}
	})

	t.Run("IfRequestJsonBody", func(t *testing.T) {
		body := `{"user": {"id": 1, "roles": ["admin", "editor"]}}`
		req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		h := NewHandlerExecutor(w, req)

		steps := []ResponseFuncConfig{
			IfRequestJsonBody("user.id", "Equal", 1, "ID_MATCH", "yes"),
			IfRequestJsonBody("user.roles[1]", "Equal", "editor", "ROLE_MATCH", "yes"),
		}
		if err := h.Execute(steps); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		if h.Variables["ID_MATCH"] != "yes" {
			t.Error("ID_MATCH not set")
		}
		if h.Variables["ROLE_MATCH"] != "yes" {
			t.Error("ROLE_MATCH not set")
		}
	})

	t.Run("IfRequestPath", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/test", nil)
		w := httptest.NewRecorder()
		h := NewHandlerExecutor(w, req)

		steps := []ResponseFuncConfig{
			IfRequestPath("Equal", "/api/v1/test", "PATH_OK", "true"),
		}
		if err := h.Execute(steps); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if h.Variables["PATH_OK"] != "true" {
			t.Error("PATH_OK not set")
		}
	})

	t.Run("IfRequestQuery", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test?q=hello", nil)
		w := httptest.NewRecorder()
		h := NewHandlerExecutor(w, req)

		steps := []ResponseFuncConfig{
			IfRequestQuery("q", "Equal", "hello", "QUERY_OK", "true"),
		}
		if err := h.Execute(steps); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if h.Variables["QUERY_OK"] != "true" {
			t.Error("QUERY_OK not set")
		}
	})
}

func TestHandlerExecutor_Generator(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h := NewHandlerExecutor(w, req)

	steps := []ResponseFuncConfig{
		GenerateRandomString(10, "R_STR"),
		GenerateRandomInt(10, 20, "R_INT"),
		GenerateRandomIntFixLength(4, "R_FIX"),
		GenerateRandomDecimal(1.0, 5.0, 2, "R_DEC"),
	}
	h.Execute(steps)

	if len(h.Variables["R_STR"].(string)) != 10 {
		t.Error("R_STR length mismatch")
	}
	rInt := h.Variables["R_INT"].(int)
	if rInt < 10 || rInt > 20 {
		t.Error("R_INT out of range")
	}
	rFix := h.Variables["R_FIX"].(int)
	if rFix < 1000 || rFix > 9999 {
		t.Error("R_FIX length mismatch")
	}
	rDec := h.Variables["R_DEC"].(float64)
	if rDec < 1.0 || rDec > 5.0 {
		t.Error("R_DEC out of range")
	}

	// Test HashedString
	h.Variables["SRC"] = "test"
	stepsHash := []ResponseFuncConfig{
		HashedString("SRC", "MD5", "HASH_MD5"),
		HashedString("SRC", "SHA256", "HASH_SHA"),
	}
	h.Execute(stepsHash)

	if len(h.Variables["HASH_MD5"].(string)) != 32 { // MD5 hex length
		t.Error("MD5 hash length mismatch")
	}
	if len(h.Variables["HASH_SHA"].(string)) != 64 { // SHA256 hex length
		t.Error("SHA256 hash length mismatch")
	}
}

func TestHandlerExecutor_DynamicVariable(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h := NewHandlerExecutor(w, req)

	h.Variables["VAL_INT"] = 123
	h.Variables["VAL_STR"] = "456"
	h.Variables["VAL_DEL"] = "trash"

	steps := []ResponseFuncConfig{
		ConvertToString("VAL_INT"),
		ConvertToInt("VAL_STR"),
		Delete("VAL_DEL"),
	}
	h.Execute(steps)

	if _, ok := h.Variables["VAL_INT"].(string); !ok {
		t.Error("VAL_INT not converted to string")
	}
	if _, ok := h.Variables["VAL_STR"].(int); !ok {
		t.Error("VAL_STR not converted to int")
	}
	if _, ok := h.Variables["VAL_DEL"]; ok {
		t.Error("VAL_DEL not deleted")
	}
}

func TestHandlerExecutor_SetupResponse(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Req", "ReqVal")
	w := httptest.NewRecorder()
	h := NewHandlerExecutor(w, req)

	h.Variables["V"] = "World"

	steps := []ResponseFuncConfig{
		SetStatusCode(201),
		SetHeader("X-Resp", "RespVal"),
		CopyHeaderFromRequest("X-Req"),
		SetJsonBody(`{"hello": "{{.V}}"}`),
	}
	h.Execute(steps)
	h.Finalize()

	resp := w.Result()
	if resp.StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
	if resp.Header.Get("X-Resp") != "RespVal" {
		t.Error("X-Resp header missing")
	}
	if resp.Header.Get("X-Req") != "ReqVal" {
		t.Error("X-Req header not copied")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if !strings.Contains(buf.String(), `"hello": "World"`) {
		t.Errorf("Body template not resolved: %s", buf.String())
	}
}

func TestResolveString(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h := NewHandlerExecutor(w, req)
	h.Variables["A"] = "AAA"

	// Direct testing of private method if needed, or via side effects
	res := h.resolveString("Val-{{.A}}")
	if res != "Val-AAA" {
		t.Errorf("Expected Val-AAA, got %s", res)
	}
}
