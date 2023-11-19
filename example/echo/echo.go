package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"	
	"sync"

	"github.com/quic-go/quic-go"
)

const addr = "localhost:4242"

const message = "foobar"

var streamMap =  make(map[quic.StreamID]quic.Stream)
var mapMutex = &sync.RWMutex{}

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	echoServer()
	// go func() { echoServer() }()

	// time.Sleep(2 * time.Second) 

	/*
	err := clientMain()
	if err != nil {
		panic(err)
	}
	*/
}

// Start a server that echos all data on the first stream opened by the client
func echoServer() {
	// fmt.Println("server started")
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		panic(err)
	}
	// fmt.Println("server listening to connections...")
	conn, err := listener.Accept(context.Background())
	if err != nil {
		panic(err)
	}
	// fmt.Println("server received client connection!")
	counter := 0
	for {
		// fmt.Println("server waiting for client stream...")
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			panic(err)
		}
		// fmt.Println("server received data with streamid:", stream.StreamID())
		go handleStream(stream)

		// Echo through the loggingWriter
		// _, err = io.Copy(loggingWriter{stream}, stream)
		/*
		streambytes, err := io.ReadAll(stream)
		if err != nil {
			panic(err)
		}
		fmt.Println(fmt.Printf("Server: Got '%s'\n", string(streambytes)))
		*/

		counter++
		// fmt.Println("num streams:", counter)
	}
}

func handleStream(stream quic.Stream) (resp string, err error) {
	mapMutex.Lock()
	streamMap[stream.StreamID()] = stream
	mapMutex.Unlock()
	// fmt.Println("Handling stream:", stream.StreamID(), stream)

	b := []byte{0} // 1 byte buffer
    var n int

    for err == nil {
        n, err = stream.Read(b)
        if n == 0 {
            continue
        }
        resp += string(b[0])
    }
	fmt.Printf("Server: Got '%s'\n", resp)
	return resp, err
}

func clientMain() error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	conn, err := quic.DialAddr(context.Background(), addr, tlsConf, nil)
	if err != nil {
		return err
	}

	fmt.Println("PADABAM")

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}

	message := "DAMMAKU"
	fmt.Printf("Client: Sending '%s'\n", message)
	_, err = stream.Write([]byte(message))
	if err != nil {
		return err
	}

	buf := make([]byte, len(message))
	_, err = io.ReadFull(stream, buf)
	if err != nil {
		return err
	}
	fmt.Printf("Client: Got '%s'\n", buf)


	stream2, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}

	message2 := "PATTAYA KELAPPU"
	fmt.Printf("Client2: Sending '%s'\n", message2)
	_, err = stream2.Write([]byte(message2))
	if err != nil {
		return err
	}

	buf2 := make([]byte, len(message2))
	_, err = io.ReadFull(stream2, buf2)
	if err != nil {
		return err
	}
	fmt.Printf("Client2: Got '%s'\n", buf2)

	return nil
}

// A wrapper for io.Writer that also logs the message.
type loggingWriter struct{ io.Writer }

func (w loggingWriter) Write(b []byte) (int, error) {
	fmt.Printf("Server: Got '%s'\n", string(b))
	return w.Writer.Write(b)
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}