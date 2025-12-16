package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	v1 "github.com/XWinterVarit/integrate_tester/pkg/v1"

	_ "github.com/sijms/go-ora/v2"
)

func main() {
	mockUrl := flag.String("mock-url", "http://localhost:9001", "Mock Server Control URL")
	flag.Parse()

	t := v1.NewTester()
	// Mock Service Port (where the actual mocked service will listen)
	mockPort := 9002

	var app *v1.AppServer
	// DB client for the test runner to manipulate DB
	var db *v1.DBClient

	t.Stage("Setup", func() {
		// Build the example app binary to ensure the latest routes (e.g., /update-json) are available
		workDir, err := os.Getwd()
		v1.AssertNoError(err)
		appPath := filepath.Join(workDir, "example_app_bin")

		buildCmd := exec.Command("go", "build", "-o", appPath, "example_app/main.go")
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			v1.Fail("Failed to build example app: %v", err)
		}

		// 1. Connect to Oracle
		dsn := "oracle://LEARN1:Welcome@localhost:1521/XE"
		db = v1.Connect("oracle", dsn)

		// Prepare Table (Oracle syntax)
		// Note: Oracle 11g/12c differences exist. We use simple NUMBER for ID.
		// Since ReplaceData provides ID, we don't need AUTO_INCREMENT/IDENTITY for this test.
		// So we just need PRIMARY KEY constraint.
		db.SetupTable("users", true, []v1.Field{
			{"id", "NUMBER PRIMARY KEY"},
			{"name", "VARCHAR2(100)"},
			{"status", "VARCHAR2(50)"},
		}, nil)

		// 2. Insert Initial Data using InsertOne (new helper for column/value pairs)
		db.InsertOne("users", []interface{}{"id", 1, "name", "alice", "status", "active"})

		// 3. Run App
		app = v1.RunAppServer(appPath, "-driver", "oracle", "-dsn", dsn, "-mock-service", fmt.Sprintf("http://localhost:%d", mockPort))
	})

	t.Stage("Success Case", func() {
		// "request test for success case"
		// 1. Update via App
		resp := v1.SendRequest("http://localhost:8080/update?id=1&status=updated")
		v1.ExpectStatusCode(resp, 200)

		// 2. Verify DB (Manipulate/Check record in DB)
		// Fetch uses QueryData which we updated to handle placeholders for Oracle
		result := db.Fetch("SELECT status FROM users WHERE id = ?", 1)
		result.ExpectCount(1)
		// Verify using simplified Expect method
		result.GetRow(0).Expect("status", "updated")

		// 3. Read via App (Should succeed now)
		resp = v1.SendRequest("http://localhost:8080/read?id=1")
		v1.ExpectJsonBody(resp, `{"id": "1", "status": "updated"}`)
	})

	t.Stage("Complex Request (POST JSON)", func() {
		requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
		newStatus := "json-updated"

		resp := v1.SendRESTRequest("http://localhost:8080/update-json",
			v1.WithMethod(http.MethodPost),
			v1.WithHeader("X-Request-ID", requestID),
			v1.WithHeaders(map[string]string{"X-Trace": "integration-example"}),
			v1.WithJSONBody(map[string]interface{}{
				"id":     "1",
				"status": newStatus,
				"meta": map[string]interface{}{
					"requested_at": time.Now().Format(time.RFC3339),
				},
			}),
		)

		v1.ExpectStatusCode(resp, 200)
		v1.ExpectHeader(resp, "Content-Type", "application/json")
		v1.ExpectJsonBodyField(resp, "id", "1")
		v1.ExpectJsonBodyField(resp, "status", newStatus)
		v1.ExpectJsonBodyField(resp, "request_id", requestID)

		// Verify DB updated from POST JSON call
		result := db.Fetch("SELECT status FROM users WHERE id = ?", 1)
		result.ExpectCount(1)
		result.GetRow(0).Expect("status", newStatus)
	})

	t.Stage("Fail Case", func() {
		// "do another request for expected fail"
		// 1. Manipulate record in DB (Set to 'bad')
		// Update uses placeholders which we updated to handle Oracle
		db.Update("users", map[string]interface{}{"status": "bad"}, "id = ?", 1)

		// 2. Request Expected Fail
		resp := v1.SendRequest("http://localhost:8080/read?id=1")
		v1.ExpectStatusCode(resp, 500)
	})

	t.Stage("Cleanup", func() {
		if app != nil {
			app.Stop()
		}
	})

	t.Stage("Mock Server Example", func() {
		// 1. Connect to the Mock Server
		// The mock server must be running separately (e.g., via docker or separate process)
		client := v1.NewDynamicMockClient(*mockUrl)

		// 2. Register a Route with Complex Logic
		// We use mockPort defined in main
		err := client.RegisterRoute(mockPort, "GET", "/mock-test", []v1.ResponseFuncConfig{
			// Generator
			v1.GenerateRandomInt(10, 50, "DISCOUNT"),

			// Conditions
			v1.IfRequestQuerySetCase("user_type", v1.ConditionEqual, "vip", "VIP"),

			// Default Response
			v1.SetStatusCode("", 200),
			v1.SetJsonBody("", `{"message": "Hello User", "discount": 0}`),
			v1.SetHeader("", "Content-Type", "application/json"),

			// VIP Response
			v1.SetStatusCode("VIP", 200),
			v1.SetJsonBody("VIP", `{"message": "Hello VIP", "discount": {{.DISCOUNT}}}`),
		})
		v1.AssertNoError(err)

		// 3. Verify Default Case
		resp := v1.SendRequest("http://localhost:8080/call-mock")
		v1.ExpectStatusCode(resp, 200)
		v1.ExpectJsonBody(resp, `{"message": "Hello User", "discount": 0}`)

		// 4. Verify VIP Case
		respVip := v1.SendRequest("http://localhost:8080/call-mock?user_type=vip")
		v1.ExpectStatusCode(respVip, 200)
		// Verify dynamic content
		if !strings.Contains(respVip.Body, "Hello VIP") {
			v1.Fail("Expected VIP message, got: %s", respVip.Body)
		}
	})

	v1.RunGUI(t)
}
