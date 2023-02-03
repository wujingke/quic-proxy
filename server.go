package main

import (
	"fmt"
	"log"
	"net"

	quic "github.com/quic-go/quic-go"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/net/context"
)

func server(c *cli.Context) error {
	// Set up our TLS
	//tlsConfig, err := configureTLS()
	tlsConfig, err := generateTLSConfig()
	if err != nil {
		fmt.Println("[!] Error grabbing TLS certs")
		return nil
	}
	quicListener, err := quic.ListenAddr(c.String("bind"), tlsConfig, nil)

	if err != nil {
		fmt.Println("[!] Error binding to UDP/443")
		return nil
	}
	fmt.Println("[+] Listening on UDP/" + c.String("bind") + " for QUIC connections")

	for {
		fmt.Println("[+] Waiting for connection...")
		session, err := quicListener.Accept(context.Background())

		if err != nil {
			fmt.Println("[!] Error accepting connection from client")
			continue
		}
		go func() {

			fmt.Printf("[*] Accepted connection from %s\n", session.RemoteAddr().String())

			stream, err := session.AcceptStream(context.Background())

			if err != nil {
				fmt.Println("[!] Error accepting stream from QUIC client")
			}

			defer session.CloseWithError(0, "Bye")
			if err != nil {
				fmt.Println("[!] Error accepting stream")
				return
			}
			defer stream.Close()

			conn2 := getTCPUpstream(c)
			if conn2 == nil {
				return
			}
			defer conn2.Close()
			fmt.Println("[+] upstream opened")

			closer := make(chan bool)
			go copy(closer, conn2, stream)
			go copy(closer, stream, conn2)
			<-closer
			log.Println("Connection complete", session.RemoteAddr())
		}()
	}
}

func getTCPUpstream(c *cli.Context) (conn net.Conn) {
	conn, err := net.Dial("tcp", c.String("addr"))
	if err != nil {
		log.Println("error dialing remote addr", err)
		return nil
	}
	return conn
}
