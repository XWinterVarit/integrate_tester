package dynamic_mock_server

// ResponseFuncConfig represents the JSON structure for a response function configuration
type ResponseFuncConfig struct {
	Group string        `json:"group"`
	Func  string        `json:"func"`
	Args  []interface{} `json:"args"`
}

// RegisterRouteRequest represents the body for /registerRoute
type RegisterRouteRequest struct {
	Port         int                  `json:"port"`
	Method       string               `json:"method"`
	Path         string               `json:"path"`
	ResponseFunc []ResponseFuncConfig `json:"responseFunc"`
}

// Constants for Response Func Groups
const (
	GroupPrepareData     = "PrepareData"
	GroupGenerator       = "Generator"
	GroupDynamicVariable = "DynamicVariable"
	GroupSetupResponse   = "SetupResponse"
)

// Constants for Response Func Names
const (
	// PrepareData
	FuncIfRequestHeader   = "IfRequestHeader"
	FuncIfRequestJsonBody = "IfRequestJsonBody"
	FuncIfRequestPath     = "IfRequestPath"
	FuncIfRequestQuery    = "IfRequestQuery"

	// Generator
	FuncGenerateRandomString       = "GenerateRandomString"
	FuncGenerateRandomInt          = "GenerateRandomInt"
	FuncGenerateRandomIntFixLength = "GenerateRandomIntFixLength"
	FuncGenerateRandomDecimal      = "GenerateRandomDecimal"
	FuncHashedString               = "HashedString"

	// DynamicVariable
	FuncConvertToString = "ConvertToString"
	FuncConvertToInt    = "ConvertToInt"
	FuncDelete          = "Delete"

	// SetupResponse
	FuncSetJsonBody           = "SetJsonBody"
	FuncSetStatusCode         = "SetStatusCode"
	FuncSetWait               = "SetWait"
	FuncSetRandomWait         = "SetRandomWait"
	FuncSetMethod             = "SetMethod"
	FuncSetHeader             = "SetHeader"
	FuncCopyHeaderFromRequest = "CopyHeaderFromRequest"
)

// Conditions
const (
	ConditionEqual = "Equal"
	// Add other conditions as needed (e.g., NotEqual, Contains, etc.) based on requirement interpretation
	// The requirement only shows "Equal" in examples, but standard comparison usually implies more.
)
