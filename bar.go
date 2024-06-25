package main

import (
	"encoding/json"
	"net"
	"os"
	"syscall"
)

func main() {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var s socket

	d := json.NewDecoder(f)
	if err := d.Decode(&s); err != nil {
		panic(err)
	}

	localAddr, err := net.ResolveTCPAddr("tcp4", s.LocalAddr)
	if err != nil {
		panic(err)
	}

	dialer := &net.Dialer{
		Control: func(network, address string, conn syscall.RawConn) error {
			return conn.Control(func(fd uintptr) {
				restore1(&s, int(fd))
			})
		},
		LocalAddr: localAddr,
		KeepAlive: -1,
	}

	conn, err := dialer.Dial("tcp4", s.RemoteAddr)
	if err != nil {
		panic(err)
	}

	sconn, err := conn.(*net.TCPConn).SyscallConn()
	if err != nil {
		panic(err)
	}

	sconn.Control(func(fd uintptr) {
		restore2(int(fd))
	})

	echo("bar", s.Revision, conn.(*net.TCPConn))
}
