package main

import (
	"flag"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	slog.Info("dippy")

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
		slog.Error("no proxy config")
		os.Exit(1)
	}

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	for _, p := range proxies {
		slog.Info("start proxy", "from", p.Addr, "to", p.Target)
		err := p.Listen()
		if err != nil {
			slog.Error("can not listen", "addr", p.Addr)
			os.Exit(1)
		}
	}

	<-shutdown

	slog.Info("shutting down")
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

			slog.Warn("accept connection error", "error", err)
			continue
		}

		slog.Info("accept connection", "local", conn.LocalAddr(), "remote", conn.RemoteAddr())
		go p.process(conn)
	}
}

func (p *proxy) process(src net.Conn) {
	defer src.Close()

	dst, err := dialer.Dial("tcp", p.Target)
	if err != nil {
		slog.Error("can not dial", "addr", p.Target, "error", err)
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
	_, err := io.Copy(p.dst, p.src)
	p.errCh <- err
}

func (p *pipeConnection) copyToSrc() {
	_, err := io.Copy(p.src, p.dst)
	p.errCh <- err
}

var dialer = net.Dialer{
	Timeout:   2 * time.Second,
	KeepAlive: 1 * time.Minute,
}

func isClosed(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "use of closed network connection")
}
