package main

import (
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var (
	targetAddr = os.Getenv("TARGET_ADDR")
	listenAddr = os.Getenv("LISTEN_ADDR")
)

func main() {
	log.Println("dippy")

	if listenAddr == "" {
		log.Fatal("LISTEN_ADDR required")
	}
	if targetAddr == "" {
		log.Fatal("TARGET_ADDR required")
	}

	log.Println("LISTEN_ADDR", listenAddr)
	log.Println("TARGET_ADDR", targetAddr)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal("can not listen on", listenAddr)
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("accept connection error;", err)
			continue
		}

		go process(conn)
	}
}

var dialer = net.Dialer{
	Timeout:   2 * time.Second,
	KeepAlive: 10 * time.Minute,
}

func process(src net.Conn) {
	log.Println("accepted")
	defer log.Println("closed")
	defer src.Close()

	dst, err := dialer.Dial("tcp", targetAddr)
	if err != nil {
		log.Println("can not dial target")
		return
	}
	defer dst.Close()

	p, b := pool.Get().([]byte), pool.Get().([]byte)
	defer pool.Put(p)
	defer pool.Put(b)
	go io.Copy(dst, src)
	io.Copy(src, dst)
}

const bufferSize = 16 * 1024

var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, bufferSize)
	},
}
