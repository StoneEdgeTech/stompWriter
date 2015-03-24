package stompWriter

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gmallard/stompngo"
)

type stompConnectioner interface {
	Send(stompngo.Headers, string) error
	Connected() bool
	Disconnect(stompngo.Headers) error
}

// A wrapper for a stomp connection to facilitate making a io.Writer interface
type StompWriter struct {
	hostname  string
	queueName string
	password  string
	port      string
	username  string

	mu sync.Mutex

	Connection        stompConnectioner
	connectionHeaders stompngo.Headers

	netCon net.Conn
}

func New(hostname, port, username, password, queueName string) (*StompWriter, error) {
	newStompWriter := StompWriter{}

	newStompWriter.hostname = hostname
	newStompWriter.port = port
	newStompWriter.username = username
	newStompWriter.password = password
	newStompWriter.queueName = queueName

	newStompWriter.mu = sync.Mutex{}

	newStompWriter.connectionHeaders = stompngo.Headers{
		"accept-version", "1.1",
		"login", username,
		"passcode", password,
		"host", hostname,
		"heart-beat", "5000,5000",
	}

	newStompWriter.Connect()

	// vvv provides hard reconnect on timeout for DNS change robustness
	go func(s *StompWriter) {
		for {
			select {
			case <-time.After(1 * time.Minute):
				s.Connect()
			}
		}
	}(&newStompWriter)

	return &newStompWriter, nil
}

func (s *StompWriter) Connect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Disconnect()
	// connect IO
	netCon, err := net.Dial("tcp", net.JoinHostPort(s.hostname, s.port))
	if err != nil {
		fmt.Printf("Couldn't connect to stomp host: %v\n", err.Error())
		return
	}
	s.netCon = netCon

	// vvv done with helper to facilitate mocking in tests
	s.Connection = getStompConnection(s.netCon, s.connectionHeaders)
}

func (s *StompWriter) Disconnect() {
	// disconnect IO
	if s.netCon != nil && s.Connection.Connected() {
		s.Connection.Disconnect(stompngo.Headers{})
	}
	if s.netCon != nil {
		s.netCon.Close()
	}
}

func (s *StompWriter) Write(payload []byte) (int, error) {
	// send message
	h := stompngo.Headers{
		"persistent", "true",
		"destination", "/queue/" + s.queueName,
		"content-type", "text/plain;charset=UTF-8",
	}
	s.mu.Lock()
	err := s.Connection.Send(h, string(payload))
	s.mu.Unlock()
	if err != nil {
		fmt.Printf("error sending stomp message: %v\n", err.Error())
	}

	return len(payload), nil
}

// vvv this is a helper to facilitate testing
var getStompConnection = func(netCon net.Conn, connectionHeaders stompngo.Headers) stompConnectioner {
	connection, err := stompngo.Connect(netCon, connectionHeaders)
	if err != nil {
		fmt.Printf("Couldn't Connect to stomp host: %v\n", err.Error())
		return nil
	}
	return connection
}
