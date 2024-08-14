package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

var addr = "127.0.0.1:2002"

// 简单实现一个http反向代理
func main() {
	rs1 := "http://127.0.0.1:2003/base"
	url1, err := url.Parse(rs1)
	if err != nil {
		log.Println(err)
	}
	rs2 := "http://127.0.0.1:2004/base"
	url2, err := url.Parse(rs2)
	if err != nil {
		log.Println(err)
	}
	urls := []*url.URL{url1, url2}

	proxy := NewSingleHostsReverseProxy(urls)
	log.Println("Starting httpproxy server at " + addr)
	log.Fatal(http.ListenAndServe(addr, proxy))
}

func NewSingleHostsReverseProxy(targets []*url.URL) *httputil.ReverseProxy {
	// copy query参数
	director := func(req *http.Request) {
		re, _ := regexp.Compile("^/dir(.*)")
		req.URL.Path = re.ReplaceAllString(req.URL.Path, "$1")

		//随机负载均衡
		targetIndex := rand.Intn(len(targets))
		target := targets[targetIndex]
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		// 目标地址+请求地址
		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		// 第一层代理设置Header头
		req.Header.Set("X-Real-Ip", req.RemoteAddr)
	}
	return &httputil.ReverseProxy{Director: director, ModifyResponse: modifyFunc}
}

func modifyFunc(resp *http.Response) error {
	if resp.StatusCode == 200 {
		return nil
	}
	oldPayload, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// 追加一部分内容
	newPayLoad := []byte("hello " + string(oldPayload))
	resp.Body = io.NopCloser(bytes.NewBuffer(newPayLoad))
	resp.ContentLength = int64(len(newPayLoad))
	resp.Header.Set("Content-Length", fmt.Sprint(len(newPayLoad)))
	return nil
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
