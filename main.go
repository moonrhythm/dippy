package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	log.Println("dippy")

	var config string
	flag.StringVar(&config, "config", os.Getenv("PROXY_CONFIG"), "proxy config")
	flag.Parse()

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

	if len(proxies) == 0 {
		log.Fatal("no proxy config")
	}

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	for _, p := range proxies {
		log.Printf("proxy %s -> %s\n", p.Addr, p.Target)
		err := p.Listen()
		if err != nil {
			log.Fatal("can not listen on", p.Addr)
		}
	}

	<-shutdown

	log.Println("shutting down")
	for _, p := range proxies {
		p.Close()
	}
}

type proxy struct {
	listener net.Listener

	Addr   string
	Target string
}

func (p *proxy) Listen() error {
	var err error
	p.listener, err = net.Listen("tcp", p.Addr)
	if err != nil {
		return err
	}

	go p.acceptLoop()
	return nil
}

func (p *proxy) Close() {
	if p.listener == nil {
		return
	}
	p.listener.Close()
}

func (p *proxy) acceptLoop() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			if isClosed(err) {
				return
			}

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
		log.Printf("can not dial %s: %s\n", p.Target, err)
		return
	}
	defer dst.Close()

	pc := pipeConnection{
		src: src,
		dst: dst,
	}
	pc.do()
}

type pipeConnection struct {
	src   net.Conn
	dst   net.Conn
	errCh chan error
}

func (p *pipeConnection) do() error {
	p.errCh = make(chan error, 2)

	go p.copyToSrc()
	go p.copyToDst()

	return <-p.errCh
}

func (p *pipeConnection) copyToDst() {
	b := pool.Get().(*[]byte)
	defer pool.Put(b)

	_, err := io.CopyBuffer(p.dst, p.src, *b)
	p.errCh <- err
}

func (p *pipeConnection) copyToSrc() {
	b := pool.Get().(*[]byte)
	defer pool.Put(b)

	_, err := io.CopyBuffer(p.src, p.dst, *b)
	p.errCh <- err
}

var dialer = net.Dialer{
	Timeout:   2 * time.Second,
	KeepAlive: 1 * time.Minute,
}

const bufferSize = 16 * 1024

var pool = sync.Pool{
	New: func() any {
		b := make([]byte, bufferSize)
		return &b
	},
}

func isClosed(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "use of closed network connection")
}
