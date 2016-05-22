package goodman

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"

	hooks "github.com/snikch/goodman/hooks"
	t "github.com/snikch/goodman/transaction"
)

const (
	defaultPort             = "61321"
	defaultMessageDelimiter = "\n"
)

type Run interface {
	RunnerInterface
	hooks.HooksServer
}

// Server is responsible for starting a server and running lifecycle callbacks.
type Server struct {
	Runner           Run
	Port             string
	MessageDelimeter []byte
	conn             net.Conn
}

// NewServer returns a new server instance with the supplied runner. If no
// runner is supplied, a new one will be created.
func NewServer(runner Run) *Server {
	if runner == nil {
		runner = NewRunner()
	}
	return &Server{
		Runner:           runner,
		Port:             defaultPort,
		MessageDelimeter: []byte(defaultMessageDelimiter),
	}
}

// Run starts the server listening for events from dredd.
func (server *Server) Run() error {
	fmt.Println("Starting")
	ln, err := net.Listen("tcp", ":"+server.Port)
	if err != nil {
		return err
	}
	fmt.Println("Accepting connection")
	conn, err := ln.Accept()
	if err != nil {
		return err
	}

	defer conn.Close()
	server.conn = conn

	for {
		// fmt.Println("Reading from connection")
		body, err := bufio.
			NewReader(conn).
			ReadString('\n')
		// fmt.Println("Read from socket")
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		body = body[:len(body)-1]
		m := &message{}
		err = json.Unmarshal([]byte(body), m)
		if err != nil {
			// fmt.Println("Unmarshal failed")
			return err
		}
		err = server.ProcessMessage(m)
		if err != nil {
			return err
		}
	}
}

// ProcessMessage handles a single event message.
func (server *Server) ProcessMessage(m *message) error {
	// fmt.Println("Processing message")
	// switch m.Event {
	// case "beforeAll":
	// 	fallthrough
	// case "afterAll":
	// 	m.transaction = &t.Transaction{}
	// 	err := json.Unmarshal(m.Data, &m.transaction)
	// 	if err != nil {
	// 		return err
	// 	}
	// default:
	m.transaction = &t.Transaction{}
	err := json.Unmarshal(m.Data, m.transaction)
	if err != nil {
		return err
	}
	// }

	switch m.Event {
	case "beforeAll":
		server.Runner.RunBeforeAll(m.transaction)
		break
	case "beforeEach":
		// before is run after beforeEach, as no separate event is fired.
		server.Runner.RunBeforeEach(m.transaction)
		server.Runner.RunBefore(m.transaction)
		break
	case "beforeEachValidation":
		// beforeValidation is run after beforeEachValidation, as no separate event
		// is fired.
		server.Runner.RunBeforeEachValidation(m.transaction)
		server.Runner.RunBeforeValidation(m.transaction)
		break
	case "afterEach":
		// after is run before afterEach as no separate event is fired.
		server.Runner.RunAfter(m.transaction)
		server.Runner.RunAfterEach(m.transaction)
		break
	case "afterAll":
		server.Runner.RunAfterAll(m.transaction)
		break
	default:
		return fmt.Errorf("Unknown event '%s'", m.Event)
	}

	// switch m.Event {
	// case "beforeAll":
	// 	fallthrough
	// case "afterAll":
	// 	return server.sendResponse(m, m.transaction)
	// default:
	return server.sendResponse(m, m.transaction)
	// }
}

// sendResponse submits the transaction(s) back to dredd.
func (server *Server) sendResponse(m *message, dataObj interface{}) error {
	data, err := json.Marshal(dataObj)
	if err != nil {
		return err
	}

	m.Data = json.RawMessage(data)
	response, err := json.Marshal(m)
	if err != nil {
		return err
	}
	server.conn.Write(response)
	server.conn.Write(server.MessageDelimeter)
	return nil
}

// message represents a single event received over the connection.
type message struct {
	UUID  string          `json:"uuid"`
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`

	transaction *t.Transaction
}
