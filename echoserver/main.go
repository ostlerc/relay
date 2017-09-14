package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

var (
	host = flag.String("h", "localhost", "host")
	port = flag.Int("p", 8080, "port")
)

func main() {
	flag.Parse()
	addr := fmt.Sprintf("%v:%v", *host, *port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	publicAddr := []byte{}
	buf := make([]byte, 100)

outer:
	for {
		n, err := conn.Read(buf)
		if err != nil {
			panic(err)
		}
		if n > 0 {
			for i := 0; i < n; i++ {
				if buf[i] == '\n' {
					publicAddr = append(publicAddr, buf[:i]...)
					break outer
				}
			}
			publicAddr = append(publicAddr, buf...)
		} else {
			time.Sleep(time.Second / 2)
		}
	}

	fmt.Printf("established relay address: %s\n", publicAddr)

	_, err = io.Copy(os.Stdout, conn)
	if err != nil {
		panic(err)
	}
}
