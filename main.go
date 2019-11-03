package main

import (
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var config = os.Getenv("PROXY_CONFIG")

func main() {
	log.Println("dippy")

	list := strings.Split(config, ",")

	// extract src and dst
	var proxies []*proxy
	for _, it := range list {
		xs := strings.Split(it, "=")
		if len(xs) != 2 {
			continue
		}
		proxies = append(proxies, &proxy{
			Addr:   xs[0],
			Target: xs[1],
		})
	}

	for _, p := range proxies {
		log.Println(p.Addr + "=" + p.Target)
		go p.Listen()
	}

	select {}
}

type proxy struct {
	Addr   string
	Target string
}

func (p *proxy) Listen() error {
	lis, err := net.Listen("tcp", p.Addr)
	if err != nil {
		log.Fatal("can not listen on", p.Addr)
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("accept connection error;", err)
			continue
		}

		go p.process(conn)
	}
}

func (p *proxy) process(src net.Conn) {
	defer src.Close()

	dst, err := dialer.Dial("tcp", p.Target)
	if err != nil {
		log.Println("can not dial target")
		return
	}
	defer dst.Close()

	s, d := pool.Get().([]byte), pool.Get().([]byte)
	defer pool.Put(s)
	defer pool.Put(d)
	go io.CopyBuffer(dst, src, s)
	io.CopyBuffer(src, dst, d)
}

var dialer = net.Dialer{
	Timeout:   2 * time.Second,
	KeepAlive: 10 * time.Minute,
}

const bufferSize = 16 * 1024

var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, bufferSize)
	},
}
