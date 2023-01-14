package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
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
// func (s *Server) Start() error {
func (s *Server) Start(intch chan os.Signal) error {
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
	<-intch
	s.quitch <- struct{}{} // TODO make quitting work right, currently <C-c> (aka SIGINT kills the server ungracefully)

	s.shutdown()

	return nil
}

// acceptLoop accept incoming connections and fires up a goroutine for each one.
func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}

		// TODO handle user/client validation before handing connection to read loop

		client := Client{conn.RemoteAddr().String(), conn}
		go s.readLoop(client)
	}
}

// readLoop continuesly reads incoming messages from one connection.
// The loop can be left if a fatal error occurs or the connection sends
// an EOF.
func (s *Server) readLoop(client Client) {
	s.clients = append(s.clients, client)

	buf := make([]byte, 2048)
	for {
		n, err := client.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("%s left\n", client.name)
				break
			}
			log.Println("read error:", err)
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
	for {
		select {
		case msg := <-s.messages:
			for _, client := range s.clients {
				writeTo(client, msg)
			}
		case <-s.quitch:
			break
		}
	}
}

// TODO make shutdown
// shutdown informs connected users that the server is going offline, collecting
// potential (non-fatal) errors along the way. If a fatal error occurs, shutdown
// will return that.
func (s *Server) shutdown() {
	shutdownMessage := Message{
		payload:   []byte("The server got shut down. Disconnected."),
		from:      Client{name: "Server"},
		arrivedAt: time.Now(),
	}

	for _, client := range s.clients {
		writeTo(client, shutdownMessage)
	}

	log.Println("Shutting Down")
}

func writeTo(client Client, msg Message) {
	fmt.Fprintf(client.conn, "%s> %s\n", msg.from.name, msg.payload)
}
