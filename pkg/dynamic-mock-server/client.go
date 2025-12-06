package dynamic_mock_server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	BaseURL string
	Client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// RegisterRoute registers a dynamic route on the mock server.
func (c *Client) RegisterRoute(port int, method, path string, responseFuncs []ResponseFuncConfig) error {
	reqBody := RegisterRouteRequest{
		Port:         port,
		Method:       method,
		Path:         path,
		ResponseFunc: responseFuncs,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := c.Client.Post(c.BaseURL+"/registerRoute", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register route: status %d", resp.StatusCode)
	}

	return nil
}

// ResetPort resets all routes for a specific port.
func (c *Client) ResetPort(port int) error {
	reqBody := map[string]int{"port": port}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := c.Client.Post(c.BaseURL+"/resetPort", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to reset port: status %d", resp.StatusCode)
	}
	return nil
}

// ResetAll resets all ports and routes.
func (c *Client) ResetAll() error {
	resp, err := c.Client.Post(c.BaseURL+"/resetAll", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to reset all: status %d", resp.StatusCode)
	}
	return nil
}

// Helper functions to create ResponseFuncConfig

func IfRequestHeader(headerName, condition, value, dynamicVar string, toBeValue interface{}) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupPrepareData,
		Func:  FuncIfRequestHeader,
		Args:  []interface{}{headerName, condition, value, dynamicVar, toBeValue},
	}
}

func IfRequestJsonBody(field, condition string, value interface{}, dynamicVar string, toBeValue interface{}) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupPrepareData,
		Func:  FuncIfRequestJsonBody,
		Args:  []interface{}{field, condition, value, dynamicVar, toBeValue},
	}
}

func IfRequestPath(condition, value, dynamicVar string, toBeValue interface{}) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupPrepareData,
		Func:  FuncIfRequestPath,
		Args:  []interface{}{condition, value, dynamicVar, toBeValue},
	}
}

func IfRequestQuery(field, condition, value, dynamicVar string, toBeValue interface{}) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupPrepareData,
		Func:  FuncIfRequestQuery,
		Args:  []interface{}{field, condition, value, dynamicVar, toBeValue},
	}
}

func ExtractRequestHeader(headerName, dynamicVar string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupPrepareData,
		Func:  FuncExtractRequestHeader,
		Args:  []interface{}{headerName, dynamicVar},
	}
}

func ExtractRequestJsonBody(field, dynamicVar string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupPrepareData,
		Func:  FuncExtractRequestJsonBody,
		Args:  []interface{}{field, dynamicVar},
	}
}

func ExtractRequestPath(dynamicVar string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupPrepareData,
		Func:  FuncExtractRequestPath,
		Args:  []interface{}{dynamicVar},
	}
}

func ExtractRequestQuery(field, dynamicVar string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupPrepareData,
		Func:  FuncExtractRequestQuery,
		Args:  []interface{}{field, dynamicVar},
	}
}

func GenerateRandomString(length int, toDynamicVariable string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupGenerator,
		Func:  FuncGenerateRandomString,
		Args:  []interface{}{length, toDynamicVariable},
	}
}

func GenerateRandomInt(min, max int, toDynamicVariable string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupGenerator,
		Func:  FuncGenerateRandomInt,
		Args:  []interface{}{min, max, toDynamicVariable},
	}
}

func GenerateRandomIntFixLength(length int, toDynamicVariable string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupGenerator,
		Func:  FuncGenerateRandomIntFixLength,
		Args:  []interface{}{length, toDynamicVariable},
	}
}

func GenerateRandomDecimal(min, max float64, maxDecimal int, toDynamicVariable string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupGenerator,
		Func:  FuncGenerateRandomDecimal,
		Args:  []interface{}{min, max, maxDecimal, toDynamicVariable},
	}
}

func HashedString(fromDynamicVariable, hashAlgorithm, toDynamicVariable string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupGenerator,
		Func:  FuncHashedString,
		Args:  []interface{}{fromDynamicVariable, hashAlgorithm, toDynamicVariable},
	}
}

func ConvertToString(dynamicVariable string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupDynamicVariable,
		Func:  FuncConvertToString,
		Args:  []interface{}{dynamicVariable},
	}
}

func ConvertToInt(dynamicVariable string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupDynamicVariable,
		Func:  FuncConvertToInt,
		Args:  []interface{}{dynamicVariable},
	}
}

func Delete(dynamicVariable string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupDynamicVariable,
		Func:  FuncDelete,
		Args:  []interface{}{dynamicVariable},
	}
}

func SetJsonBody(jsonBody string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupSetupResponse,
		Func:  FuncSetJsonBody,
		Args:  []interface{}{jsonBody},
	}
}

func SetStatusCode(statusCode int) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupSetupResponse,
		Func:  FuncSetStatusCode,
		Args:  []interface{}{statusCode},
	}
}

func SetWait(timeoutMs int) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupSetupResponse,
		Func:  FuncSetWait,
		Args:  []interface{}{timeoutMs},
	}
}

func SetRandomWait(minMs, maxMs int) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupSetupResponse,
		Func:  FuncSetRandomWait,
		Args:  []interface{}{minMs, maxMs},
	}
}

func SetMethod(method string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupSetupResponse,
		Func:  FuncSetMethod,
		Args:  []interface{}{method},
	}
}

func SetHeader(key, value string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupSetupResponse,
		Func:  FuncSetHeader,
		Args:  []interface{}{key, value},
	}
}

func CopyHeaderFromRequest(key string) ResponseFuncConfig {
	return ResponseFuncConfig{
		Group: GroupSetupResponse,
		Func:  FuncCopyHeaderFromRequest,
		Args:  []interface{}{key},
	}
}
