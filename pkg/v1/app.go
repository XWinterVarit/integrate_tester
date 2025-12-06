package v1

import (
	"log"
	"os"
	"os/exec"
)

// AppServer represents a running application server.
type AppServer struct {
	cmd *exec.Cmd
}

// RunAppServer runs the application server.
func RunAppServer(path string, args ...string) *AppServer {
	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("[App] Starting Server: %s %v", path, args)
	if err := cmd.Start(); err != nil {
		log.Printf("[App] Failed to start server: %v", err)
		return nil
	}

	return &AppServer{cmd: cmd}
}

// Stop stops the application server.
func (s *AppServer) Stop() {
	if s.cmd != nil && s.cmd.Process != nil {
		log.Println("[App] Stopping Server")
		s.cmd.Process.Kill()
		s.cmd.Wait() // release resources
	}
}
