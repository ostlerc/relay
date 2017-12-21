package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var (
	relay = flag.String("r", ":8080", "relay address")
)

func main() {
	flag.Parse()
	conn, err := net.Dial("tcp", *relay)
	if err != nil {
		panic(err)
	}

	sigint := make(chan os.Signal)
	signal.Notify(sigint, syscall.SIGINT)
	go func() {
		<-sigint
		conn.Close()
		os.Exit(0)
	}()

	r := bufio.NewReader(conn)
	for {
		dat, err := r.ReadBytes('\n')
		clientAddr := string(dat[:len(dat)-1])
		fmt.Println("new client", clientAddr)

		client, err := net.Dial("tcp", clientAddr)
		if err != nil {
			panic(err)
		}

		go echo(client)
	}
}

func echo(conn net.Conn) error {
	_, err := io.Copy(conn, conn)
	if err != nil {
		fmt.Println("Err echoing", err)
	}

	closeErr := conn.Close()
	if closeErr != nil {
		fmt.Println("Err closing", closeErr)
	}
	return err
}
