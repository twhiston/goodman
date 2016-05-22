package goodman

import (
	"bufio"
	"net"
	"testing"
)

func TestSendingServerMessages(t *testing.T) {
	server := NewServer(NewRunner())

	// ch := make(chan int)
	go func() {
		err := server.Run()
		if err != nil {
			t.Fatalf("Dredd hooks server failed to start with error %s", err.Error())
		}
		// <-ch
	}()

	messages := []struct {
		Payload []byte
	}{
		// TODO: Figure why I could not use `` quoted string
		{
			Payload: []byte("{\"uuid\":\"1234-abcd\",\"event\":\"beforeEach\",\"data\":{}}\n"),
		},
		{
			Payload: []byte("{\"uuid\":\"2234-abcd\",\"event\":\"beforeEachValidation\",\"data\":{}}\n"),
		},
		{
			Payload: []byte("{\"uuid\":\"2234-abcd\",\"event\":\"afterEach\",\"data\":{}}\n"),
		},
		{
			Payload: []byte("{\"uuid\":\"2234-abcd\",\"event\":\"beforeAll\",\"data\":{}}\n"),
		},
		{
			Payload: []byte("{\"uuid\":\"2234-abcd\",\"event\":\"afterAll\",\"data\":{}}\n"),
		},
	}

	conn, err := net.Dial("tcp", "localhost:61321")

	if err != nil {
		t.Fatalf("Client connection to dredd hooks server failed")
	}

	for _, v := range messages {

		_, err := conn.Write(v.Payload)

		if err != nil {
			t.Errorf("Sending message %s failed with error %s", string(v.Payload), err.Error())
		}

		body, err := bufio.NewReader(conn).ReadString(byte('\n'))
		if body != string(v.Payload) {
			t.Errorf("Body of %s does not match the payload of %s", body, string(v.Payload))
		}
	}
}

// Setting runner....
// Have RPC Server that hooks connect to and pass Runner through.

// func TestRunHooks(t *testing.T) {
// 	ch := make(chan int)
// 	go func() {
// 		RunHooksServer()
// 		<-ch
// 	}()

// 	client, err := rpc.DialHTTP("tcp", "localhost:1234")
// 	if err != nil {
// 		log.Fatal("dialing:", err)
// 	}
// 	args := *new(Test)
// 	reply := new(bool)
// 	err = client.Call("HooksServer.Test", args, reply)
// 	fmt.Println(*reply)
// 	if err != nil {
// 		log.Fatal("arith error:", err)
// 	}
// }
