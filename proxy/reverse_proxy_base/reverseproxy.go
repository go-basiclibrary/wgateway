package main

import (
	"bufio"
	"log"
	"net/http"
	"net/url"
)

var (
	proxyAddr = "http://127.0.0.1:2003"
	port      = "2002"
)

func handler(w http.ResponseWriter, req *http.Request) {
	// 获取代理地址
	proxy, err := url.Parse(proxyAddr)
	req.URL.Scheme = proxy.Scheme
	req.URL.Host = proxy.Host

	// 封装请求到下游
	transport := http.DefaultTransport
	resp, err := transport.RoundTrip(req)
	if err != nil {
		log.Print(err)
		return
	}

	// 把下游请求返回给上游
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	defer resp.Body.Close()
	bufio.NewReader(resp.Body).WriteTo(w)
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Start serving on port" + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
