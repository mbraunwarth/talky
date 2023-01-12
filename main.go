package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

// Server struct which handles incoming connections and overall application state.
type Server struct {
	ln     net.Listener
	quitch chan struct{}
	errs   []Error
}

// Error type alias.
type Error error

// NewServer returns a fresh Server, note that the servers listener is not
// properly setup yet.
func NewServer() *Server {
	return &Server{
		quitch: make(chan struct{}),
	}
}

// The Start function sets up the servers listener and returns possible errors.
// It is also responsible for shutdown handling, hence a nil error means a graceful
// server shutdown.
func (s *Server) Start() error {
	// TODO outsource magic strings and numbers to args or even config struct
	ln, err := net.Listen("tcp", "localhost:2000")
	if err != nil {
		return err
	}
	defer ln.Close()

	s.ln = ln

	go s.acceptLoop()

	// shutdown once the server received a quit signal
	<-s.quitch
	return s.shutdown()
}

// acceptLoop accept incoming connections and fires up a goroutine for each one.
func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go s.readLoop(conn)
	}
}

// readLoop continuesly reads incoming messages from one connection.
// The loop can be left if a fatal error occurs or the connection sends
// an EOF.
func (s *Server) readLoop(conn net.Conn) {
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("%s left\n", conn.RemoteAddr())
				break
			}
			log.Println(err)
			continue
		}

		msg := buf[:n]
		fmt.Printf("received msg from %s: %s\n", conn.RemoteAddr(), msg)
	}
}

// shutdown informs connected users that the server is going offline, collecting
// potential (non-fatal) errors along the way. If a fatal error occurs, shutdown
// will return that.
func (s *Server) shutdown() error {
	// Pseudo Go
	//for c := range s.conns {
	//	if err := c.Write("shutting down the server"); err != nil {
	//		s.errs = append(s.errs, Error{err})
	//	}
	//}
	return nil
}

func main() {
	server := NewServer()
	// TODO **Multiple** logs of `use of closed network connection` error
	//		in stdout, meaning someone tries to use s.ln somewhere after shutdown

	// remove comment to test quit channel
	//go func() {
	//	time.Sleep(10 * time.Second)
	//	server.quitch <- struct{}{}
	//}()
	//------------------------------------

	log.Fatal(server.Start())

}
