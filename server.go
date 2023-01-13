package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// Server struct which handles incoming connections and overall application state.
type Server struct {
	ln net.Listener

	quitch   chan struct{}
	messages chan Message

	errs    []Error
	clients []Client
}

// A Message from any kind of client.
type Message struct {
	from      Client
	arrivedAt time.Time
	payload   []byte
}

// Client struct defines the properties of a client connected to the server.
type Client struct {
	name string
	conn net.Conn
}

// Error type alias.
type Error error

// NewServer returns a fresh Server, note that the servers listener is not
// properly setup yet.
func NewServer() *Server {
	return &Server{
		quitch:   make(chan struct{}),
		messages: make(chan Message, 1000),
		errs:     make([]Error, 0),
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
	// TODO move close to shutdown??
	defer ln.Close()

	go s.acceptLoop()
	go s.broadcast()

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

		// TODO handle user/client validation before handing connection to read loop

		client := Client{conn.RemoteAddr().String(), conn}
		go s.readLoop(client)
	}
}

// TODO handing read loop a validated Client instead of raw net.Conn
// readLoop continuesly reads incoming messages from one connection.
// The loop can be left if a fatal error occurs or the connection sends
// an EOF.
func (s *Server) readLoop(client Client) {
	// TODO add client to s.clients
	s.clients = append(s.clients, client)

	buf := make([]byte, 2048)
	for {
		n, err := client.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("%s left\n", client.name)
				break
			}
			log.Println(err)
			continue
		}

		msg := Message{
			from:      client,
			arrivedAt: time.Now(),
			payload:   buf[:n-1], // cut off newline for ncat and telnet clients
		}

		s.messages <- msg

		log.Printf("received msg from %s: %s\n", client.name, msg.payload)
	}
}

// broadcast incoming messages to every connected client.
func (s *Server) broadcast() {
	for msg := range s.messages {
		for _, client := range s.clients {
			fmt.Fprintf(client.conn, "%s> %s\n", msg.from.name, msg.payload)
		}
	}
}

// TODO make shutdown
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
