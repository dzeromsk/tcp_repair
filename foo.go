package main

import (
	"net"
)

func main() {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 8443,
	})
	if err != nil {
		panic(err)
	}

	println("listen", ln.Addr().String())

	conn, err := ln.AcceptTCP()
	if err != nil {
		panic(err)
	}

	echo("foo", 0, conn)
}
