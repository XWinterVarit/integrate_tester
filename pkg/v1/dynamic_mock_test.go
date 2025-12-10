package v1

import (
	"testing"

	dm "github.com/XWinterVarit/integrate_tester/pkg/dynamic-mock-server"
)

func TestDynamicMockAliases(t *testing.T) {
	// Test a few functions to ensure they are aliased correctly and return expected config

	// Test SetStatusCode
	conf := SetStatusCode("Case1", 200)
	if conf.Group != dm.GroupSetupResponse {
		t.Errorf("SetStatusCode Group incorrect: %s", conf.Group)
	}
	if conf.Func != dm.FuncSetStatusCode {
		t.Errorf("SetStatusCode Func incorrect: %s", conf.Func)
	}
	if len(conf.Args) != 2 || conf.Args[0] != "Case1" || conf.Args[1] != 200 {
		t.Errorf("SetStatusCode Args incorrect: %v", conf.Args)
	}

	// Test IfRequestHeader
	conf = IfRequestHeader("H", ConditionEqual, "V", "VAR", "Val")
	if conf.Group != dm.GroupPrepareData {
		t.Errorf("IfRequestHeader Group incorrect")
	}
	if conf.Args[1] != ConditionEqual {
		t.Errorf("IfRequestHeader Condition incorrect")
	}

	// Test Client Creation
	client := NewDynamicMockClient("http://localhost:8888")
	if client.BaseURL != "http://localhost:8888" {
		t.Errorf("Client BaseURL incorrect")
	}
}
