#MockStomp
Library for adding mocking capabilities around STOMP connections.

#Compatibility
This library was created with the sole purpose of mocking https://github.com/gmallard/stompngo.  Mocking any other STOMP library may work, but is not intentional.

#Usage
To use this library, a slight change in the implementation of stompngo is necessary.

In the primary project
```go

type StompConnector interface {
	Send(stompngo.Headers, string) error
	Connected() bool
	Disconnect(stompngo.Headers) error
}

var (
	...

	getStompConnection = func(c net.Conn, h stompngo.Headers) (StompConnector, error) {
		return stompngo.Connect(c, h)
	}
)

func ConnectToBrokerFunc(){
	...

	// Establish stomp connection
	stomp, err := getStompConnection(b.conn, b.connHeaders)

	...
}
```

In the tests
```go
import (
	"github.com/gmallard/stompngo"
	"github.com/stoneedgetech/mockstomp"
)

func TestStompFunctionality(t *testing.T){
	getStompConnection = func(c net.Conn, h stompngo.Headers) (StompConnector, error) {
		return &mockstomp.MockStompConnection{}, nil
	}

	...
	//Test connect function
}
```
