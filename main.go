package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"sync"
)

var (
	host = flag.String("h", "localhost", "host")
	port = flag.Int("p", 8080, "port")
)

type Server struct {
	host string
	port int

	portmu         sync.Mutex
	relayPortStart int
}

func (s *Server) Serve() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", s.host, s.port))
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
	fmt.Println("New relay server connection", conn.RemoteAddr())
	p := s.NextPort()
	addr := fmt.Sprintf("%v:%v", s.host, p)
	l, err := net.Listen("tcp", addr) // TODO: some resiliency around finding an address to connect to
	if err != nil {
		fmt.Println("failed to listen to addr", addr, err)
		return
	}

	_, err = conn.Write([]byte(addr + "\n"))
	if err != nil {
		fmt.Println("Failed to send relay host", err)
		_ = conn.Close()
	}

	r := &relay{
		host:     s.host,
		port:     p,
		listener: l,
		client:   conn,
	}
	go func() {
		err = r.Serve()
		if err != nil {
			fmt.Printf("serve failed%v\n\n%#v\n", err, r)
		}
		fmt.Println("Finished client serving")
	}()
}

func (s *Server) NextPort() int {
	s.portmu.Lock()
	if s.relayPortStart == 0 {
		s.relayPortStart = *port
	}

	s.relayPortStart++
	defer s.portmu.Unlock()
	return s.relayPortStart
}

type relay struct {
	host string
	port int

	listener net.Listener
	client   net.Conn
	done     bool
}

func (r *relay) Serve() error {
	defer r.listener.Close()
	for {
		conn, err := r.listener.Accept()
		if err != nil {
			if r.done {
				return nil
			}
			return err
		}

		go func(conn net.Conn) {
			fmt.Println("New relay client connection", conn.RemoteAddr())
			_, err = io.Copy(r.client, conn)
			if err != nil {
				fmt.Println("Closing relay client and server connection", conn.RemoteAddr(), err)
				r.listener.Close()
			}
			fmt.Println("Closing relay client connection", conn.RemoteAddr())
			conn.Close()
		}(conn)
	}
}

func main() {
	flag.Parse()
	s := &Server{
		host: *host,
		port: *port,
	}
	err := s.Serve()
	if err != nil {
		panic(err)
	}
}
