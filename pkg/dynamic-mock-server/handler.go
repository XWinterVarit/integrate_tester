package dynamic_mock_server

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// HandlerExecutor executes the response functions
type HandlerExecutor struct {
	Variables      map[string]interface{}
	Request        *http.Request
	ParsedBody     interface{}
	ResponseWriter http.ResponseWriter

	// Response State
	StatusCode int
	Body       string
	Headers    map[string]string
	FixedDelay time.Duration
	RandomWait [2]int // min, max
	ActiveCase string
}

func NewHandlerExecutor(w http.ResponseWriter, r *http.Request) *HandlerExecutor {
	return &HandlerExecutor{
		Variables:      make(map[string]interface{}),
		Request:        r,
		ResponseWriter: w,
		StatusCode:     200,
		Headers:        make(map[string]string),
	}
}

func (h *HandlerExecutor) Execute(funcs []ResponseFuncConfig) error {
	// Pre-parse body if needed
	if h.Request.Body != nil {
		bodyBytes, _ := io.ReadAll(h.Request.Body)
		h.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore for reading if needed
		if len(bodyBytes) > 0 {
			json.Unmarshal(bodyBytes, &h.ParsedBody)
		}
	}

	for _, f := range funcs {
		if err := h.runFunc(f); err != nil {
			return err
		}
	}

	return nil
}

func (h *HandlerExecutor) Finalize() {
	// Apply delays
	if h.FixedDelay > 0 {
		time.Sleep(h.FixedDelay)
	}
	if h.RandomWait[1] > 0 {
		min := h.RandomWait[0]
		max := h.RandomWait[1]
		if max > min {
			sleepTime := time.Duration(rand.Intn(max-min)+min) * time.Millisecond
			time.Sleep(sleepTime)
		}
	}

	// Apply headers
	for k, v := range h.Headers {
		h.ResponseWriter.Header().Set(k, v)
	}

	// Write status
	h.ResponseWriter.WriteHeader(h.StatusCode)

	// Write body
	// Apply template to body one last time if it contains variables?
	// The requirement says SetJsonBody takes a template string.
	// So h.Body likely already stores the template string.
	// We should execute it now.
	finalBody := h.resolveString(h.Body)
	h.ResponseWriter.Write([]byte(finalBody))
}

func (h *HandlerExecutor) runFunc(f ResponseFuncConfig) error {
	switch f.Group {
	case GroupPrepareData:
		return h.handlePrepareData(f)
	case GroupGenerator:
		return h.handleGenerator(f)
	case GroupDynamicVariable:
		return h.handleDynamicVariable(f)
	case GroupSetupResponse:
		return h.handleSetupResponse(f)
	}
	return nil
}

// Helper to resolve templates in strings
func (h *HandlerExecutor) resolveString(s string) string {
	if !strings.Contains(s, "{{") {
		return s
	}
	t, err := template.New("tmpl").Parse(s)
	if err != nil {
		return s // Return raw if parse fails
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, h.Variables); err != nil {
		return s // Return raw if execute fails
	}
	return buf.String()
}

func (h *HandlerExecutor) resolveArg(arg interface{}) interface{} {
	if s, ok := arg.(string); ok {
		return h.resolveString(s)
	}
	return arg
}

func (h *HandlerExecutor) handlePrepareData(f ResponseFuncConfig) error {
	args := f.Args
	var condition string
	var expectedVal interface{}
	var targetVar string
	var toBeVal interface{}
	var actualVal interface{}

	switch f.Func {
	case FuncIfRequestHeader:
		if len(args) < 5 {
			return nil
		}
		condition = fmt.Sprintf("%v", args[1])
		expectedVal = h.resolveArg(args[2])
		targetVar = fmt.Sprintf("%v", args[3])
		toBeVal = h.resolveArg(args[4])

		headerName := fmt.Sprintf("%v", args[0])
		actualVal = h.Request.Header.Get(headerName)
	case FuncIfRequestJsonBody:
		if len(args) < 5 {
			return nil
		}
		condition = fmt.Sprintf("%v", args[1])
		expectedVal = h.resolveArg(args[2])
		targetVar = fmt.Sprintf("%v", args[3])
		toBeVal = h.resolveArg(args[4])

		fieldPath := fmt.Sprintf("%v", args[0])
		actualVal = h.getJSONPath(fieldPath)
	case FuncIfRequestPath:
		if len(args) < 4 {
			return nil
		}
		condition = fmt.Sprintf("%v", args[0])
		expectedVal = h.resolveArg(args[1])
		targetVar = fmt.Sprintf("%v", args[2])
		toBeVal = h.resolveArg(args[3])
		actualVal = h.Request.URL.Path
	case FuncIfRequestQuery:
		if len(args) < 5 {
			return nil
		}
		condition = fmt.Sprintf("%v", args[1])
		expectedVal = h.resolveArg(args[2])
		targetVar = fmt.Sprintf("%v", args[3])
		toBeVal = h.resolveArg(args[4])

		queryField := fmt.Sprintf("%v", args[0])
		actualVal = h.Request.URL.Query().Get(queryField)

	case FuncIfRequestHeaderSetCase:
		if len(args) < 4 {
			return nil
		}
		condition = fmt.Sprintf("%v", args[1])
		expectedVal = h.resolveArg(args[2])
		caseStr := fmt.Sprintf("%v", args[3])

		headerName := fmt.Sprintf("%v", args[0])
		actualVal = h.Request.Header.Get(headerName)
		if h.checkCondition(actualVal, condition, expectedVal) {
			h.ActiveCase = caseStr
		}
		return nil

	case FuncIfRequestJsonBodySetCase:
		if len(args) < 4 {
			return nil
		}
		condition = fmt.Sprintf("%v", args[1])
		expectedVal = h.resolveArg(args[2])
		caseStr := fmt.Sprintf("%v", args[3])

		fieldPath := fmt.Sprintf("%v", args[0])
		actualVal = h.getJSONPath(fieldPath)
		if h.checkCondition(actualVal, condition, expectedVal) {
			h.ActiveCase = caseStr
		}
		return nil

	case FuncIfRequestPathSetCase:
		if len(args) < 3 {
			return nil
		}
		condition = fmt.Sprintf("%v", args[0])
		expectedVal = h.resolveArg(args[1])
		caseStr := fmt.Sprintf("%v", args[2])
		actualVal = h.Request.URL.Path
		if h.checkCondition(actualVal, condition, expectedVal) {
			h.ActiveCase = caseStr
		}
		return nil

	case FuncIfRequestQuerySetCase:
		if len(args) < 4 {
			return nil
		}
		condition = fmt.Sprintf("%v", args[1])
		expectedVal = h.resolveArg(args[2])
		caseStr := fmt.Sprintf("%v", args[3])

		queryField := fmt.Sprintf("%v", args[0])
		actualVal = h.Request.URL.Query().Get(queryField)
		if h.checkCondition(actualVal, condition, expectedVal) {
			h.ActiveCase = caseStr
		}
		return nil

	case FuncExtractRequestHeader:
		if len(args) < 2 {
			return nil
		}
		headerName := fmt.Sprintf("%v", args[0])
		targetVar := fmt.Sprintf("%v", args[1])
		h.Variables[targetVar] = h.Request.Header.Get(headerName)
		return nil

	case FuncExtractRequestJsonBody:
		if len(args) < 2 {
			return nil
		}
		fieldPath := fmt.Sprintf("%v", args[0])
		targetVar := fmt.Sprintf("%v", args[1])
		val := h.getJSONPath(fieldPath)
		if val != nil {
			h.Variables[targetVar] = val
		}
		return nil

	case FuncExtractRequestPath:
		if len(args) < 1 {
			return nil
		}
		targetVar := fmt.Sprintf("%v", args[0])
		h.Variables[targetVar] = h.Request.URL.Path
		return nil

	case FuncExtractRequestQuery:
		if len(args) < 2 {
			return nil
		}
		queryField := fmt.Sprintf("%v", args[0])
		targetVar := fmt.Sprintf("%v", args[1])
		h.Variables[targetVar] = h.Request.URL.Query().Get(queryField)
		return nil
	}

	if h.checkCondition(actualVal, condition, expectedVal) {
		h.Variables[targetVar] = toBeVal
	}

	return nil
}

func (h *HandlerExecutor) checkCondition(actual interface{}, cond string, expected interface{}) bool {
	// Simple string comparison for now, as most inputs are HTTP parts
	actStr := fmt.Sprintf("%v", actual)
	expStr := fmt.Sprintf("%v", expected)

	switch cond {
	case ConditionEqual:
		return actStr == expStr
	}
	return false
}

func (h *HandlerExecutor) getJSONPath(path string) interface{} {
	if h.ParsedBody == nil {
		return nil
	}
	parts := strings.Split(path, ".")
	var current interface{} = h.ParsedBody

	for _, part := range parts {
		// handle array index like a[0]
		key := part
		idx := -1

		if strings.Contains(part, "[") && strings.HasSuffix(part, "]") {
			// Extract key and index
			idxStart := strings.Index(part, "[")
			key = part[:idxStart]
			idxStr := part[idxStart+1 : len(part)-1]
			idx, _ = strconv.Atoi(idxStr)
		}

		// Access map
		m, ok := current.(map[string]interface{})
		if !ok {
			// Maybe current is just the array? e.g. path "0.field"
			// But JSON root is usually object.
			return nil
		}

		val, exists := m[key]
		if !exists {
			return nil
		}
		current = val

		// Access array if index present
		if idx >= 0 {
			arr, ok := current.([]interface{})
			if !ok || idx >= len(arr) {
				return nil
			}
			current = arr[idx]
		}
	}
	return current
}

func (h *HandlerExecutor) handleGenerator(f ResponseFuncConfig) error {
	args := f.Args
	switch f.Func {
	case FuncGenerateRandomString:
		length := int(toFloat(args[0]))
		targetVar := fmt.Sprintf("%v", args[1])
		h.Variables[targetVar] = randomString(length)
	case FuncGenerateRandomInt:
		min := int(toFloat(args[0]))
		max := int(toFloat(args[1]))
		targetVar := fmt.Sprintf("%v", args[2])
		h.Variables[targetVar] = rand.Intn(max-min+1) + min
	case FuncGenerateRandomIntFixLength:
		length := int(toFloat(args[0]))
		targetVar := fmt.Sprintf("%v", args[1])
		// Not perfect but works for simple case
		min := int(1 * pow10(length-1))
		max := int(1*pow10(length) - 1)
		h.Variables[targetVar] = rand.Intn(max-min+1) + min
	case FuncGenerateRandomDecimal:
		min := toFloat(args[0])
		max := toFloat(args[1])
		// maxDecimal := int(toFloat(args[2])) // unused in simple implementation
		targetVar := fmt.Sprintf("%v", args[3])
		val := min + rand.Float64()*(max-min)
		h.Variables[targetVar] = val
	case FuncHashedString:
		fromVar := fmt.Sprintf("%v", args[0])
		algo := fmt.Sprintf("%v", args[1])
		targetVar := fmt.Sprintf("%v", args[2])

		val := fmt.Sprintf("%v", h.Variables[fromVar])
		var hash string
		if algo == "MD5" {
			sum := md5.Sum([]byte(val))
			hash = hex.EncodeToString(sum[:])
		} else if algo == "SHA256" {
			sum := sha256.Sum256([]byte(val))
			hash = hex.EncodeToString(sum[:])
		}
		h.Variables[targetVar] = hash
	}
	return nil
}

func (h *HandlerExecutor) handleDynamicVariable(f ResponseFuncConfig) error {
	args := f.Args
	targetVar := fmt.Sprintf("%v", args[0])

	switch f.Func {
	case FuncConvertToString:
		if v, ok := h.Variables[targetVar]; ok {
			h.Variables[targetVar] = fmt.Sprintf("%v", v)
		}
	case FuncConvertToInt:
		if v, ok := h.Variables[targetVar]; ok {
			h.Variables[targetVar] = int(toFloat(v))
		}
	case FuncDelete:
		delete(h.Variables, targetVar)
	}
	return nil
}

func (h *HandlerExecutor) handleSetupResponse(f ResponseFuncConfig) error {
	args := f.Args
	if len(args) == 0 {
		return nil
	}
	caseStr := fmt.Sprintf("%v", args[0])
	// If ActiveCase is "", it matches "" (default)
	// If ActiveCase is "CaseA", it matches "CaseA"
	if caseStr != h.ActiveCase {
		return nil
	}

	switch f.Func {
	case FuncSetJsonBody:
		h.Body = fmt.Sprintf("%v", args[1])
	case FuncSetStatusCode:
		h.StatusCode = int(toFloat(args[1]))
	case FuncSetWait:
		h.FixedDelay = time.Duration(toFloat(args[1])) * time.Millisecond
	case FuncSetRandomWait:
		h.RandomWait[0] = int(toFloat(args[1]))
		h.RandomWait[1] = int(toFloat(args[2]))
	case FuncSetMethod:
		// Usually response doesn't set method, maybe this is for asserting?
		// Or maybe it's mimicking? The req says "SetMethod".
		// Unclear usage for response, ignoring for now or logging.
	case FuncSetHeader:
		key := fmt.Sprintf("%v", args[1])
		val := h.resolveString(fmt.Sprintf("%v", args[2]))
		h.Headers[key] = val
	case FuncCopyHeaderFromRequest:
		key := fmt.Sprintf("%v", args[1])
		val := h.Request.Header.Get(key)
		if val != "" {
			h.Headers[key] = val
		}
	}
	return nil
}

// Utils

func toFloat(i interface{}) float64 {
	switch v := i.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	return 0
}

func pow10(n int) float64 {
	r := 1.0
	for i := 0; i < n; i++ {
		r *= 10
	}
	return r
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
