package v1

import (
	dm "github.com/XWinterVarit/integrate_tester/pkg/dynamic-mock-server"
)

// DynamicMockClient is a wrapper around the dynamic mock server client.
type DynamicMockClient struct {
	*dm.Client
}

// ResponseFuncConfig aliases the configuration struct from dynamic-mock-server.
type ResponseFuncConfig = dm.ResponseFuncConfig

// Constants for Conditions
const (
	ConditionEqual              = dm.ConditionEqual
	ConditionNotEqual           = dm.ConditionNotEqual
	ConditionContains           = dm.ConditionContains
	ConditionNotContains        = dm.ConditionNotContains
	ConditionStartsWith         = dm.ConditionStartsWith
	ConditionEndsWith           = dm.ConditionEndsWith
	ConditionGreaterThan        = dm.ConditionGreaterThan
	ConditionLessThan           = dm.ConditionLessThan
	ConditionGreaterThanOrEqual = dm.ConditionGreaterThanOrEqual
	ConditionLessThanOrEqual    = dm.ConditionLessThanOrEqual
)

// NewDynamicMockClient creates a new client for an existing dynamic mock server.
// controlURL is the base URL of the mock controller (e.g., "http://localhost:8888").
func NewDynamicMockClient(controlURL string) *DynamicMockClient {
	client := dm.NewClient(controlURL)
	return &DynamicMockClient{Client: client}
}

// Generator and Condition Functions Aliases

var (
	IfRequestHeader          = dm.IfRequestHeader
	IfRequestJsonBody        = dm.IfRequestJsonBody
	IfRequestPath            = dm.IfRequestPath
	IfRequestQuery           = dm.IfRequestQuery
	IfRequestHeaderSetCase   = dm.IfRequestHeaderSetCase
	IfRequestJsonBodySetCase = dm.IfRequestJsonBodySetCase
	IfRequestPathSetCase     = dm.IfRequestPathSetCase
	IfRequestQuerySetCase    = dm.IfRequestQuerySetCase

	IfDynamicVariable        = dm.IfDynamicVariable
	IfDynamicVariableSetCase = dm.IfDynamicVariableSetCase

	IfRequestJsonArrayLength         = dm.IfRequestJsonArrayLength
	IfRequestJsonArrayLengthSetCase  = dm.IfRequestJsonArrayLengthSetCase
	IfRequestJsonObjectLength        = dm.IfRequestJsonObjectLength
	IfRequestJsonObjectLengthSetCase = dm.IfRequestJsonObjectLengthSetCase
	IfRequestJsonType                = dm.IfRequestJsonType
	IfRequestJsonTypeSetCase         = dm.IfRequestJsonTypeSetCase

	ExtractRequestHeader   = dm.ExtractRequestHeader
	ExtractRequestJsonBody = dm.ExtractRequestJsonBody
	ExtractRequestPath     = dm.ExtractRequestPath
	ExtractRequestQuery    = dm.ExtractRequestQuery

	GenerateRandomString       = dm.GenerateRandomString
	GenerateRandomInt          = dm.GenerateRandomInt
	GenerateRandomIntFixLength = dm.GenerateRandomIntFixLength
	GenerateRandomDecimal      = dm.GenerateRandomDecimal
	HashedString               = dm.HashedString

	ConvertToString     = dm.ConvertToString
	ConvertToInt        = dm.ConvertToInt
	DynamicVarSubstring = dm.DynamicVarSubstring
	DynamicVarJoin      = dm.DynamicVarJoin
	Delete              = dm.Delete

	SetJsonBody           = dm.SetJsonBody
	SetStatusCode         = dm.SetStatusCode
	SetWait               = dm.SetWait
	SetRandomWait         = dm.SetRandomWait
	SetMethod             = dm.SetMethod
	SetHeader             = dm.SetHeader
	CopyHeaderFromRequest = dm.CopyHeaderFromRequest
)
