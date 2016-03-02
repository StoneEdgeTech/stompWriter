package mockstomp

/*
 * provides a MockStompConnection struct with accompaning functions to implement the
 * stomp interface that will record what gets called. See mock_test.go
 * for usage examples, but in general, it looks like this:
 * > mockStompConnectionInstance.Send(headers,message)
 * >
 */

import (
	"fmt"
	"github.com/gmallard/stompngo"
)

type MockStompMessage struct {
	Order   int
	Headers stompngo.Headers
	Message string
}

type MockStompConnection struct {
	Messages         chan MockStompMessage
	NumMessages      int
	DisconnectCalled bool
	Subscription     <-chan stompngo.MessageData
	subscription     chan stompngo.MessageData
}

func (m *MockStompConnection) Clear() {
	m.Messages = make(chan MockStompMessage, 1000)
	m.DisconnectCalled = false
}

func (m *MockStompConnection) Disconnect(stompngo.Headers) error {
	m.DisconnectCalled = true
	return nil
}

func (m MockStompConnection) Connected() bool {
	return true
}

func (m MockStompConnection) PutToSubscribe(msg stompngo.MessageData) {
	m.subscription <- msg
}

func New() *MockStompConnection {
	msgs := make(chan MockStompMessage, 1000)
	s := make(chan stompngo.MessageData, 1000)
	return &MockStompConnection{
		Messages:     msgs,
		subscription: s,
	}
}

func (m *MockStompConnection) Send(headers stompngo.Headers, message string) error {
	// check for protocol

	// check for destination header
	if headers.Value("destination") == "" {
		return fmt.Errorf("No destination header, cannot send.")
	}

	// save for later
	msg := MockStompMessage{len(m.Messages), headers, message}
	m.Messages <- msg

	m.NumMessages++

	return nil
}

func (m *MockStompConnection) Subscribe(stompngo.Headers) (<-chan stompngo.MessageData, error) {
	m.Subscription = (<-chan stompngo.MessageData)(m.subscription)
	return m.Subscription, nil
}

func (m *MockStompConnection) Unsubscribe(stompngo.Headers) error {
	m.Subscription = nil
	return nil
}
