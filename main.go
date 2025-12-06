package main

import (
	v1 "integrate_tester/pkg/v1"
)

func main() {
	// Create the tester instance
	t := v1.NewTester()

	// Variables shared across stages
	// In a real scenario, these would be initialized in the "Setup" stage.
	var mockServer *v1.MockServer
	// var dbClient *v1.DBClient
	// var appServer *v1.AppServer

	// Define "Setup" stage
	t.Stage("Setup", func() {
		// DB Connection (Example)
		// dbClient = v1.Connect("postgres", "user=... dbname=...")
		// dbClient.SetupTable("table1", true, []v1.Field{{"id", "TEXT"}, {"name", "TEXT"}}, nil)
		// dbClient.ReplaceData("table1", []interface{}{"1", "a"})

		// Run Mock Server
		// Port 8081 for mock
		mockServer = v1.RunMockServer("8081", map[string]v1.MockHandlerFunc{
			"/a": func(req v1.Request) v1.Response {
				return v1.Response{StatusCode: 200, Body: `{"status":"a"}`, Header: map[string]string{"Content-Type": "application/json"}}
			},
			"/b": func(req v1.Request) v1.Response {
				return v1.NewResponse(200, `{"status":"b"}`)
			},
		})

		// Run App Server (Example)
		// appServer = v1.RunAppServer("./app", "-config", "test.conf")
	})

	// Define "Case 1" stage
	t.Stage("Case 1", func() {
		// In a real test, we would call the App Server, which calls the Mock Server.
		// Here we call the Mock Server directly for demonstration.
		resp := v1.SendRequest("http://localhost:8081/a")

		v1.ExpectHeader(resp, "Content-Type", "application/json")
		v1.ExpectJsonBody(resp, `{"status":"a"}`)
	})

	// Define "Case 2" stage
	t.Stage("Case 2", func() {
		// Update Mock Server behavior
		v1.UpdateMockServer(mockServer, map[string]v1.MockHandlerFunc{
			"/a": func(req v1.Request) v1.Response {
				return v1.NewResponse(200, `{"status":"c"}`)
			},
		})

		resp := v1.SendRequest("http://localhost:8081/a")
		v1.ExpectJsonBody(resp, `{"status":"c"}`)
	})

	t.Stage("Cleanup", func() {
		if mockServer != nil {
			mockServer.Stop()
		}
		// if appServer != nil { appServer.Stop() }
	})

	// Run the GUI to control the test execution
	v1.RunGUI(t)
}
