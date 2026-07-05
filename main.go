package main

import (
	"fmt"
	"net"

	"github.com/lukekeum/go-rtmp/handshake"
)

func main() {
	l, err := net.Listen("tcp", ":1935")
	if err != nil {
		fmt.Println("Error while listening to port 1935: ", err)
		return
	}
	defer l.Close()

	fmt.Println("Server started on port 1935")

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error while accpet request: ", err)
			continue
		}
		fmt.Printf("%+v\n", c)
		go handleConn(c)
	}
}

func handleConn(c net.Conn) {
	// defer c.Close()

	if err := handshake.Connect(c); err != nil {
		
	}

}