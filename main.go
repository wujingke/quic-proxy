package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"

	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name: "server",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "bind", Value: "0.0.0.0:4242"},
					&cli.StringFlag{Name: "addr", Value: "127.0.0.1:3306"},
				},
				Action: server,
			},
			{
				Name: "client",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "bind", Value: "127.0.0.1:4242"},
					&cli.StringFlag{Name: "addr", Value: "127.0.0.1:3306"},
				},
				Action: client,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{tlsCert},
		NextProtos:         []string{"quic-proxy"},
	}, nil
}

func copy(closer chan bool, dst io.Writer, src io.Reader) {
	//r := io.TeeReader(src, dst)
	//_, _ = io.Copy(os.Stdout, r)
	fmt.Println("[+] Copying data!")
	io.Copy(dst, src)
	fmt.Println("[+] done Copying data")
	closer <- true // connection is closed, send signal to stop proxy
}
