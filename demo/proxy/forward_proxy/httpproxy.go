package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// local proxy

type Pxy struct{}

// imp Handler
func (p *Pxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Received request %s %s %s\n", req.Method, req.Host, req.RemoteAddr)
	transport := http.DefaultTransport
	// step1 浅拷贝对象,然后就再新增属性数据
	outRep := new(http.Request)
	*outRep = *req
	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := outRep.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		outRep.Header.Set("X-Forwarded-For", clientIP)
	}

	// step2 请求下游
	res, err := transport.RoundTrip(outRep)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	// step3 把下游请求内容返回给上游
	for key, value := range res.Header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}
	rw.WriteHeader(res.StatusCode)
	io.Copy(rw, res.Body)
	res.Body.Close()
}

func main() {
	fmt.Println("Serve on :8080")
	http.Handle("/", &Pxy{})
	http.ListenAndServe("0.0.0.0:8080", nil)
}
