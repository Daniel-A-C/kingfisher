package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Message is sent from the kf CLI or deep-link handler to the running Kingfisher app.
type Message struct {
	Action string `json:"action"` // "focus" | "open"
	Path   string `json:"path,omitempty"`
	Slug   string `json:"slug,omitempty"`
}

// SocketPath returns the Unix domain socket path for single-instance IPC.
func SocketPath() string {
	if custom := strings.TrimSpace(os.Getenv("KINGFISHER_SOCKET")); custom != "" {
		return custom
	}
	return filepath.Join(os.TempDir(), "kingfisher.sock")
}

// Server listens on a Unix domain socket and dispatches incoming Messages.
type Server struct {
	socketPath string
	ln         net.Listener
	handler    func(Message)
}

// StartServer binds the Unix socket (cleaning up stale sockets) and begins serving.
func StartServer(socketPath string, handler func(Message)) (*Server, error) {
	// If a socket file exists, check whether an active server is listening
	if _, err := os.Stat(socketPath); err == nil {
		conn, dialErr := net.DialTimeout("unix", socketPath, 150*time.Millisecond)
		if dialErr == nil {
			_ = conn.Close()
			return nil, fmt.Errorf("another Kingfisher instance is already listening on %s", socketPath)
		}
		// Stale socket file — safe to remove
		_ = os.Remove(socketPath)
	}

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("listen unix %s: %w", socketPath, err)
	}

	s := &Server{
		socketPath: socketPath,
		ln:         ln,
		handler:    handler,
	}
	go s.acceptLoop()
	return s, nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err == nil && s.handler != nil {
			s.handler(msg)
		}
	}
}

// Close shuts down the IPC server and removes the socket file.
func (s *Server) Close() error {
	err := s.ln.Close()
	_ = os.Remove(s.socketPath)
	return err
}

// Send delivers a Message to a running Kingfisher instance.
func Send(socketPath string, msg Message) error {
	conn, err := net.DialTimeout("unix", socketPath, 300*time.Millisecond)
	if err != nil {
		return err
	}
	defer conn.Close()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_ = conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.Write(data)
	return err
}
