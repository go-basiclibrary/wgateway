package main

import (
	"context"
	"net"
)

var (
	addr = ":2002"
)

type tcpHandler struct {
}

func (t *tcpHandler) ServeTCP(ctx context.Context, conn net.Conn) {
	conn.Write([]byte("tcpHandler123\n"))
}

type tcpH struct {
}

func (t *tcpH) ServeTCP(ctx context.Context, conn net.Conn) {
	conn.Write([]byte("tcpHandler123\n"))
}

func main() {
	// tcp服务器测试
	//log.Println("Starting tcp server at " + addr)
	//tcpSrv := tcp_proxy.TcpServer{Addr: addr, Handler: &tcpHandler{}}
	//tcpSrv.ListenAndServe()
}
