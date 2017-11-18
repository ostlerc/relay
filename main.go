package main

import (
	"flag"
	"fmt"
	"io"
	"net"
)

var (
	host = flag.String("h", "localhost", "host")
	port = flag.Int("p", 8080, "port")
)

type Server struct {
	addr string
}

func (s *Server) Serve() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go s.Accept(conn)
	}
}

func (s *Server) Accept(conn net.Conn) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("failed to create relay connection", err)
		return
	}

	fmt.Println("New relay server connection", l.Addr())

	r := &relay{
		relayClient: l,
		conn:        conn,
	}
	go func() {
		err = r.Serve()
		if err != nil {
			fmt.Printf("serve failed%v\n", err)
		}
		fmt.Println("Finished client serving")
	}()
}

type relay struct {
	relayClient net.Listener
	conn        net.Conn
}

func (r *relay) Serve() error {
	defer r.relayClient.Close()
	defer r.conn.Close()
	defer func() {
		fmt.Println("Closed relay client server")
	}()
	for {
		conn, err := r.relayClient.Accept()
		if err != nil {
			return err
		}

		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			return err
		}

		// tell the relay client about the clients relay port
		_, err = r.conn.Write([]byte(listener.Addr().String() + "\n"))
		if err != nil {
			return err
		}
		clientConn, err := listener.Accept()
		if err != nil {
			return err
		}
		defer listener.Close()
		fmt.Println("New relay client connection", clientConn.RemoteAddr())

		go r.copyIo(conn, clientConn, conn)
		go r.copyIo(conn, conn, clientConn)
	}
}

func (r *relay) copyIo(conn net.Conn, w io.Writer, reader io.Reader) {
	_, err := io.Copy(w, reader)
	if err != nil {
		fmt.Println("Closing relay client and server connection", conn.RemoteAddr(), err)
		r.relayClient.Close()
	}
	fmt.Println("Closing relay client connection", conn.RemoteAddr())
	conn.Close()
}

func main() {
	flag.Parse()
	s := &Server{addr: fmt.Sprintf("%s:%d", *host, *port)}
	if err := s.Serve(); err != nil {
		panic(err)
	}
}
