package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"golang.org/x/sys/unix"
)

const (
	TCP_NO_QUEUE = iota
	TCP_RECV_QUEUE
	TCP_SEND_QUEUE
	TCP_QUEUES_NR
)

type socket struct {
	Revision   int
	LocalAddr  string
	RemoteAddr string
	InputSeq   int
	OutputSeq  int
}

const filename = "socket.dat"

func echo(prompt string, revision int, conn *net.TCPConn) {
	defer conn.Close()

	println("connected")
	fmt.Fprintf(conn, "%s%d> hello\n", prompt, revision)

	go io.Copy(conn, os.Stdin)

	s := bufio.NewScanner(conn)
	for s.Scan() {
		fmt.Fprintf(conn, "%s%d> %s\n", prompt, revision, s.Text())

		if s.Text() == "quit" {
			sconn, err := conn.SyscallConn()
			if err != nil {
				panic(err)
			}
			s := socket{
				Revision:   revision + 1,
				LocalAddr:  conn.LocalAddr().String(),
				RemoteAddr: conn.RemoteAddr().String(),
			}
			sconn.Control(func(fd uintptr) {
				save(&s, int(fd))
			})
			println("save", filename)
			writeFile(filename, &s)
			return
		}
	}
}

func save(s *socket, fd int) {
	// enable repair mode
	err := unix.SetsockoptInt(fd, unix.SOL_TCP, unix.TCP_REPAIR, 1)
	if err != nil {
		panic(err)
	}

	// use recv sequence number
	err = unix.SetsockoptInt(fd, unix.SOL_TCP, unix.TCP_REPAIR_QUEUE, TCP_RECV_QUEUE)
	if err != nil {
		panic(err)
	}
	s.InputSeq, err = unix.GetsockoptInt(fd, unix.SOL_TCP, unix.TCP_QUEUE_SEQ)
	if err != nil {
		panic(err)
	}

	// get send sequence number
	err = unix.SetsockoptInt(fd, unix.SOL_TCP, unix.TCP_REPAIR_QUEUE, TCP_SEND_QUEUE)
	if err != nil {
		panic(err)
	}
	s.OutputSeq, err = unix.GetsockoptInt(fd, unix.SOL_TCP, unix.TCP_QUEUE_SEQ)
	if err != nil {
		panic(err)
	}
}

func restore1(s *socket, fd int) {
	// enable repair mode
	err := unix.SetsockoptInt(fd, unix.SOL_TCP, unix.TCP_REPAIR, 1)
	if err != nil {
		panic(err)
	}
	// set recv sequence number
	err = unix.SetsockoptInt(fd, unix.SOL_TCP, unix.TCP_REPAIR_QUEUE, TCP_RECV_QUEUE)
	if err != nil {
		panic(err)
	}
	err = unix.SetsockoptInt(fd, unix.SOL_TCP, unix.TCP_QUEUE_SEQ, s.InputSeq)
	if err != nil {
		panic(err)
	}
	// set send sequence number
	err = unix.SetsockoptInt(fd, unix.SOL_TCP, unix.TCP_REPAIR_QUEUE, TCP_SEND_QUEUE)
	if err != nil {
		panic(err)
	}
	err = unix.SetsockoptInt(fd, unix.SOL_TCP, unix.TCP_QUEUE_SEQ, s.OutputSeq)
	if err != nil {
		panic(err)
	}
}

func restore2(fd int) {
	// disable repair mode
	err := unix.SetsockoptInt(fd, unix.SOL_TCP, unix.TCP_REPAIR, 0)
	if err != nil {
		panic(err)
	}
}

func writeFile(name string, s *socket) {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	if err := e.Encode(s); err != nil {
		panic(err)
	}
}
