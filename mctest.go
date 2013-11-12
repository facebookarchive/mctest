// Package mctest provides standalone instances of memcache suitable for use in
// tests.
package mctest

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/ParsePlatform/go.freeport"
	"github.com/bradfitz/gomemcache/memcache"
)

// Fatalf is satisfied by testing.T or testing.B.
type Fatalf interface {
	Fatalf(format string, args ...interface{})
}

// Server is a unique instance of a memcached.
type Server struct {
	Port int
	T    Fatalf
	cmd  *exec.Cmd
}

// Start the server, this will return once the server has been started.
func (s *Server) Start() {
	port, err := freeport.Get()
	if err != nil {
		s.T.Fatalf(err.Error())
	}
	s.Port = port

	ports := fmt.Sprint(port)
	s.cmd = exec.Command("memcached", "-p", ports, "-U", ports)
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr
	if err := s.cmd.Start(); err != nil {
		s.T.Fatalf(err.Error())
	}

	// Wait until TCP socket is active to ensure we don't progress until the
	// server is ready to accept.
	for {
		if c, err := net.Dial("tcp", s.Addr()); err == nil {
			c.Close()
			break
		}
	}
}

// Addr for the server.
func (s *Server) Addr() string {
	return fmt.Sprintf("localhost:%d", s.Port)
}

// Stop the server.
func (s *Server) Stop() {
	s.cmd.Process.Kill()
}

// Client returns a memcache.Client connected to the underlying server.
func (s *Server) Client() *memcache.Client {
	return memcache.New(s.Addr())
}

// NewStartedServer creates a new server starts it.
func NewStartedServer(t Fatalf) *Server {
	s := &Server{T: t}
	s.Start()
	return s
}
