package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"quic-go/quic-go"
)

const addr = "localhost:4242"

const message = "foobar"

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	err := clientMain()
	if err != nil {
		panic(err)
	}
}

func clientMain() error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}

	fmt.Println("client up!")

	conn, err := quic.DialAddr(context.Background(), addr, tlsConf, nil)
	if err != nil {
		return err
	}

	// fmt.Println("client sent connection!")

	// fmt.Println("client to send streams...")

	// time.Sleep(1*time.Second)

	stream, err := conn.OpenStream()
	if err != nil {
		return err
	}

	// fmt.Println("client opened one stream with id:", stream.StreamID())

	// time.Sleep(1*time.Second)

	message := "MESSAGE1 "
	_, err = stream.Write([]byte(message))
	if err != nil {
		return err
	}
	// fmt.Printf("Client: Sending '%s' on stream id: %d \n", message, stream.StreamID())

	// stream.Close()

	/*
	buf := make([]byte, len(message))
	_, err = io.ReadFull(stream, buf)
	if err != nil {
		return err
	}
	fmt.Printf("Client: Got '%s'\n", buf)
	*/

	// time.Sleep(1*time.Second)

	stream2, err := conn.OpenStream()
	if err != nil {
		return err
	}

	// fmt.Println("client opened another stream with id:", stream2.StreamID())

	// time.Sleep(1 * time.Second)

	message2 := "MESSAGE2 " 
	_, err = stream2.Write([]byte(message2))
	if err != nil {
		return err
	}
	// fmt.Printf("Client: Sending '%s' on stream id: %d \n", message2, stream2.StreamID())


	stream3, err := conn.OpenStream()
	if err != nil {
		return err
	}
	message3 := "MESSAGE3 "
	_, err = stream3.Write([]byte(message3))
	if err != nil {
		return err
	}



	

	message = "MESSAGE1 "
	_, err = stream.Write([]byte(message))
	if err != nil {
		return err
	}
	// fmt.Printf("Client: Sending '%s' on stream id: %d \n", message, stream.StreamID())


	message2 = "MESSAGE2 "
	_, err = stream2.Write([]byte(message2))
	if err != nil {
		return err
	}
	// fmt.Printf("Client: Sending '%s' on stream id: %d \n", message2, stream2.StreamID())

	message3 = "MESSAGE3 "
	_, err = stream3.Write([]byte(message3))
	if err != nil {
		return err
	}

	message = "MESSAGE1 "
	_, err = stream.Write([]byte(message))
	if err != nil {
		return err
	}
	// fmt.Printf("Client: Sending '%s' on stream id: %d \n", message, stream.StreamID())


	message2 = "MESSAGE2 "
	_, err = stream2.Write([]byte(message2))
	if err != nil {
		return err
	}
	// fmt.Printf("Client: Sending '%s' on stream id: %d \n", message2, stream2.StreamID())

	message3 = "MESSAGE3 "
	_, err = stream3.Write([]byte(message3))
	if err != nil {
		return err
	}

	message = "MESSAGE1 "
	_, err = stream.Write([]byte(message))
	if err != nil {
		return err
	}
	// fmt.Printf("Client: Sending '%s' on stream id: %d \n", message, stream.StreamID())


	message2 = "MESSAGE2 "
	_, err = stream2.Write([]byte(message2))
	if err != nil {
		return err
	}
	// fmt.Printf("Client: Sending '%s' on stream id: %d \n", message2, stream2.StreamID())

	message3 = "MESSAGE3 "
	_, err = stream3.Write([]byte(message3))
	if err != nil {
		return err
	}


	message = "MESSAGE1"
	_, err = stream.Write([]byte(message))
	if err != nil {
		return err
	}
	// fmt.Printf("Client: Sending '%s' on stream id: %d \n", message, stream.StreamID())

	stream.Close()




	stream4, err := conn.OpenStream()
	if err != nil {
		return err
	}
	message4 := "MESSAGE4 "
	_, err = stream4.Write([]byte(message4))
	if err != nil {
		return err
	}


	// time.Sleep(1 * time.Second)

	message2 = "MESSAGE2"
	_, err = stream2.Write([]byte(message2))
	if err != nil {
		return err
	}
	// fmt.Printf("Client: Sending '%s' on stream id: %d \n", message2, stream2.StreamID())

	stream2.Close()

	
	message3 = "MESSAGE3"
	_, err = stream3.Write([]byte(message3))
	if err != nil {
		return err
	}

	stream3.Close()
	

	message4 = "MESSAGE4 "
	_, err = stream4.Write([]byte(message4))
	if err != nil {
		return err
	}

	message4 = "MESSAGE4 "
	_, err = stream4.Write([]byte(message4))
	if err != nil {
		return err
	}

	message4 = "MESSAGE4 "
	_, err = stream4.Write([]byte(message4))
	if err != nil {
		return err
	}

	message4 = "MESSAGE4"
	_, err = stream4.Write([]byte(message4))
	if err != nil {
		return err
	}


	time.Sleep(1 * time.Second)

	stream4.Close()

	time.Sleep(1 * time.Second)

	/*
	buf2 := make([]byte, len(message2))
	_, err = io.ReadFull(stream2, buf2)
	if err != nil {
		return err
	}
	fmt.Printf("Client2: Got '%s'\n", buf2)
	*/

	return nil
}