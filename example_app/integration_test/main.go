package main

import (
	"flag"
	"fmt"
	"os"

	v1 "integrate_tester/pkg/v1"

	_ "github.com/sijms/go-ora/v2"
)

func main() {
	mockUrl := flag.String("mock-url", "http://localhost:8888", "Mock Server Control URL")
	flag.Parse()

	t := v1.NewTester()
	var app *v1.AppServer
	// DB client for the test runner to manipulate DB
	var db *v1.DBClient

	t.Stage("Setup", func() {
		appPath := "../../example_app_bin"
		if _, err := os.Stat(appPath); os.IsNotExist(err) {
			if _, err := os.Stat("example_app_bin"); err == nil {
				appPath = "./example_app_bin"
			}
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

		// 2. Insert Initial Data
		db.ReplaceData("users", []interface{}{1, "alice", "active"})

		// 3. Run App
		app = v1.RunAppServer(appPath, "-driver", "oracle", "-dsn", dsn)
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

		// 2. Register a Route
		// We'll use port 9002 for the mock service to avoid conflict with the main app (8080) and default control port (9001)
		mockPort := 9002
		err := client.RegisterRoute(mockPort, "GET", "/mock-test", []v1.ResponseFuncConfig{
			v1.SetStatusCode("", 200),
			v1.SetJsonBody("", `{"message": "hello from mock"}`),
			v1.SetHeader("", "Content-Type", "application/json"),
		})
		v1.AssertNoError(err)

		// 3. Verify the Mock
		// Send a request to the mocked endpoint
		resp := v1.SendRequest(fmt.Sprintf("http://localhost:%d/mock-test", mockPort))

		v1.ExpectStatusCode(resp, 200)
		v1.ExpectJsonBody(resp, `{"message": "hello from mock"}`)
	})

	v1.RunGUI(t)
}
