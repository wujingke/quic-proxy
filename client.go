package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"

	quic "github.com/quic-go/quic-go"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/net/context"
)

func client(c *cli.Context) error {
	listener, err := net.Listen("tcp", c.String("bind"))
	log.Printf("Listening at %q...", c.String("bind"))
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		log.Println("New connection", conn.RemoteAddr())
		if err != nil {
			log.Println("error accepting connection", err)
			continue
		}

		go func() {
			defer conn.Close()

			conn2 := getQuicUpstream(c)
			if conn2 == nil {
				return
			}
			defer conn2.CloseWithError(0, "")
			stream, err := conn2.OpenStreamSync(context.Background())
			if err != nil {
				fmt.Println("[!] Error opening stream")
				return
			}
			defer stream.Close()

			doneCh := make(chan bool)
			go copy(doneCh, stream, conn)
			go copy(doneCh, conn, stream)
			<-doneCh
			log.Println("Connection complete", conn.RemoteAddr())
		}()
	}
}

func getQuicUpstream(c *cli.Context) (conn quic.Connection) {
	// Set up our TLS
	//tlsConfig, err := configureTLS()

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-proxy"},
	}
	conn, err := quic.DialAddr(c.String("addr"), tlsConf, nil)
	if err != nil {
		panic(err)
	}
	return conn

}
