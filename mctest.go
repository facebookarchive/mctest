// Package mctest provides standalone instances of memcache suitable for use in
// tests.
package mctest

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/facebookgo/freeport"
	"github.com/facebookgo/waitout"
)

var serverListening = []byte("server listening")

// Fatalf is satisfied by testing.T or testing.B.
type Fatalf interface {
	Fatalf(format string, args ...interface{})
}

// Server is a unique instance of a memcached.
type Server struct {
	Port        int
	StopTimeout time.Duration
	T           Fatalf
	cmd         *exec.Cmd
}

// Start the server, this will return once the server has been started.
func (s *Server) Start() {
	port, err := freeport.Get()
	if err != nil {
		s.T.Fatalf(err.Error())
	}
	s.Port = port

	waiter := waitout.New(serverListening)
	s.cmd = exec.Command("memcached", "-vv", "-l", s.Addr())
	if testing.Verbose() {
		s.cmd.Stdout = os.Stdout
		s.cmd.Stderr = io.MultiWriter(os.Stderr, waiter)
	} else {
		s.cmd.Stderr = waiter
	}
	if err := s.cmd.Start(); err != nil {
		s.T.Fatalf(err.Error())
	}
	waiter.Wait()

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
	return fmt.Sprintf("127.0.0.1:%d", s.Port)
}

// Stop the server.
func (s *Server) Stop() {
	fin := make(chan struct{})
	go func() {
		defer close(fin)
		s.cmd.Process.Kill()
	}()
	select {
	case <-fin:
	case <-time.After(s.StopTimeout):
	}
}

// Client returns a memcache.Client connected to the underlying server.
func (s *Server) Client() *memcache.Client {
	c := memcache.New(s.Addr())
	c.Timeout = time.Second
	return c
}

// NewStartedServer creates a new server starts it.
func NewStartedServer(t Fatalf) *Server {
	s := &Server{
		T:           t,
		StopTimeout: 15 * time.Second,
	}
	s.Start()
	return s
}
