/*
   Copyright 2015 Albus <albus@shaheng.me>.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"bufio"
	//"bytes"
	"crypto/tls"
	//"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	//"reflect"
	"encoding/json"
	"runtime"
	"strings"
	"time"
)

var CipherSuites = map[string]uint16{
	"TLS_RSA_WITH_RC4_128_SHA":                uint16(0x0005),
	"TLS_RSA_WITH_3DES_EDE_CBC_SHA":           uint16(0x000a),
	"TLS_RSA_WITH_AES_128_CBC_SHA":            uint16(0x002f),
	"TLS_RSA_WITH_AES_256_CBC_SHA":            uint16(0x0035),
	"TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":        uint16(0xc007),
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":    uint16(0xc009),
	"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":    uint16(0xc00a),
	"TLS_ECDHE_RSA_WITH_RC4_128_SHA":          uint16(0xc011),
	"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":     uint16(0xc012),
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":      uint16(0xc013),
	"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":      uint16(0xc014),
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   uint16(0xc02f),
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": uint16(0xc02b),

	// TLS_FALLBACK_SCSV isn't a standard cipher suite but an indicator
	// that the client is doing version fallback. See
	// https://tools.ietf.org/html/draft-ietf-tls-downgrade-scsv-00.
	"TLS_FALLBACK_SCSV": uint16(0x5600),
}

var (
	help        *bool   = flag.Bool("h", false, "show help")
	url         *string = flag.String("u", "https://0.0.0.0:28080/", "server url")
	concurrency *int    = flag.Int("c", 1, "concurrency")
	reqNum      *int    = flag.Int("r", 0, "total requests per connection")
	dur         *int    = flag.Int("t", 0, "timelimit (msec)")
	withReq     *bool   = flag.Bool("k", false, "send request after handshake on the keep-alive connection")
	cipherSuite *string = flag.String("s", "TLS_RSA_WITH_RC4_128_SHA", "cipher suite")
	header      *string = flag.String("H", "", "request header")
)

var (
	proto     string
	host      string
	port      string
	address   string
	path      string
	swithHttp bool = false
	interval  int
	network   string          = "tcp"
	servers   map[string]bool = make(map[string]bool)
	headers   []string
)

type Reporter struct {
	Server              string
	Hostname            string
	Port                string
	Path                string
	Headers             string
	ContentLength       int64
	Concurrency         int
	TimeTaken           int64
	TimeDur             int64
	TotalRequest        int
	FailedRequest       int
	RequestPerSecond    int
	ConnectionPerSecond int
	Non2XXCode          int
}

func (r *Reporter) Printer() error {
	report := fmt.Sprintf("Server Software:%s\nServer Hostname:%s\nServer Port:%s\n\nRequest Headers:\n%s\n\nDocument Path:%s\nDocument Length:%d\n\nConcurrency:%d\nTime Duration:%dms\nAvg Time Taken:%dms\n\nComplete Requests:%d\nFailed Request:%d\n\nRequest Per Second:%d\nConnections Per Second:%d\n\nNon2XXCode:%d\n\n", r.Server, r.Hostname, r.Port, r.Headers, r.Path, r.ContentLength, r.Concurrency, r.TimeDur, r.TimeTaken/1000/int64(r.TotalRequest), r.TotalRequest, r.FailedRequest, r.RequestPerSecond, r.ConnectionPerSecond, r.Non2XXCode)
	fmt.Println(report)
	return nil
}

func printHelp() {
	fmt.Println("Usage: iBenchmark [options]")
	flag.PrintDefaults()
	fmt.Printf("\ncihper suite:\n")
	for k := range CipherSuites {
		fmt.Printf("  %s\n", k)
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			printHelp()
		}
	}()
	interval = 3600
	flag.Parse()
	if *help {
		printHelp()
		return
	}
	proto = (*url)[:strings.Index(*url, ":")]
	if proto == "http" {
		swithHttp = true
	}
	host = (*url)[strings.Index(*url, "//")+2 : strings.LastIndexAny(*url, ":")]
	port = (*url)[strings.LastIndex(*url, ":")+1 : strings.LastIndex(*url, "/")]
	address = host + ":" + port
	path = (*url)[strings.LastIndex(*url, "/"):]
	if host == "" || port == "" || path == "" || proto == "" {
		printHelp()
		return
	}
	runtime.GOMAXPROCS(8)

	timeout := time.Duration(*dur) * time.Millisecond
	finChan := make([]chan bool, *concurrency)

	// number of connections to crypto server cluster
	reporter := new(Reporter)
	reporter.Concurrency = *concurrency
	reporter.Hostname = host
	reporter.Port = port
	reporter.Path = path

	fmt.Println("benchmark start ")
	// start workers
	start := time.Now()
	for i := 0; i < *concurrency; i = i + 1 {
		finChan[i] = make(chan bool)
		go worker(*reqNum, timeout, reporter, finChan[i], i)
	}

	// wait for finish
	for i := 0; i < *concurrency; i = i + 1 {
		switch {
		case <-(finChan[i]):
			continue
		}
	}
	duration := time.Since(start).Nanoseconds() / (1000 * 1000)
	reporter.TimeDur = duration
	if *withReq {
		reporter.RequestPerSecond = int(float64(reporter.TotalRequest) / (float64(reporter.TimeDur) / 1000))
		reporter.ConnectionPerSecond = 0
	} else {
		reporter.ConnectionPerSecond = int(float64(reporter.TotalRequest) / (float64(reporter.TimeDur) / 1000))
		reporter.RequestPerSecond = 0
	}
	var server string
	for key, _ := range servers {
		server = fmt.Sprintf("%s %s", server, key)
	}
	reporter.Server = server
	reporter.Headers = *header
	time.Sleep(1 * time.Second)
	reporter.Printer()
}

func (r *Reporter) GetResponse(conn *net.Conn) error {
	var resp *http.Response
	var err error
	cipher := CipherSuites[*cipherSuite]
	procStart := time.Now()
	r.TotalRequest += 1
	if !swithHttp {
		if !*withReq {
			resp, err = HTTPSGet(cipher)
		} else {
			resp, err = HTTPSGet_KeepAlive(cipher, conn)
		}

	} else {
		if !*withReq {
			resp, err = HTTPGet()
		} else {
			resp, err = HTTPGet_KeepAlive(conn)
		}
	}
	if err != nil {
		fmt.Println(fmt.Sprintf("HTTP(S) GET ERROR %v", err))
		r.FailedRequest += 1
	}
	if resp != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			r.Non2XXCode += 1
		}
		r.ContentLength = resp.ContentLength
		for _, server := range resp.Header["Server"] {
			if !servers[server] {
				servers[server] = true
			}
		}
	}
	resp.Body.Close()
	end := time.Now()
	elapse := end.Sub(procStart).Nanoseconds() / 1000
	r.TimeTaken += elapse
	return err
}

func worker(reqNum int, timeout time.Duration, reporter *Reporter, finChan chan bool, index int) {
	end_time := time.After(timeout)
	var conn net.Conn

	defer func() {
		if conn != nil {
			conn.Close()
		}

	}()
	if *dur != 0 {
		for {
			select {
			case <-end_time:
				finChan <- true
				return
			default:
				err := reporter.GetResponse(&conn)
				if err != nil {
					fmt.Println(fmt.Sprintf("[ERROR]:%s", err))
					conn = nil
				}

			}
		}

	} else {
		for i := 0; i < reqNum; i++ {
			err := reporter.GetResponse(&conn)
			if err != nil {
				fmt.Println(fmt.Sprintf("[ERROR]:%s", err))
				//conn = nil
			}
		}
		finChan <- true
		return
	}

}

func HTTPSGet(cipherSuite uint16) (*http.Response, error) {
	// create tls config
	config := tls.Config{
		InsecureSkipVerify:     true,
		SessionTicketsDisabled: true,
		CipherSuites:           []uint16{cipherSuite},
	}
	// connect to tls server
	conn, err := tls.Dial(network, address, &config)
	if err != nil {
		fmt.Errorf("client: dial: %s", err)
		return nil, err
	}
	if *withReq {
		return SendQuery(conn)
	} else {
		return nil, nil
	}
}

func HTTPSGet_KeepAlive(cipherSuite uint16, conn *net.Conn) (*http.Response, error) {
	// create tls config
	config := tls.Config{
		InsecureSkipVerify:     true,
		SessionTicketsDisabled: true,
		CipherSuites:           []uint16{cipherSuite},
	}
	var err error
	// connect to tls server
	if *conn == nil {
		*conn, err = tls.Dial(network, address, &config)
		if err != nil {
			fmt.Errorf("client: dial: %s", err)
			return nil, err
		}

	}
	resp, err := SendQuery(*conn)
	return resp, err
}

func HTTPGet() (*http.Response, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		fmt.Errorf("client: dial: %s", err)
		return nil, err
	}

	if *withReq {
		return SendQuery(conn)
	} else {
		return nil, nil
	}
}

func HTTPGet_KeepAlive(conn *net.Conn) (*http.Response, error) {
	var err error
	if *conn == nil {
		*conn, err = net.Dial(network, address)
		if err != nil {
			fmt.Errorf("client: dial: %s", err)
			return nil, err
		}

	}
	resp, err := SendQuery(*conn)
	return resp, err
}

func SendQuery(conn net.Conn) (*http.Response, error) {
	var resp *http.Response
	var err error
	// send message
	//message := "GET /shaheng.html HTTP/1.1\r\nHost: baike.baidu.com\r\n\r\n"
	var message string
	var temp string
	if *header != "" {
		json.Unmarshal([]byte(*header), &headers)
		for _, h := range headers {
			temp = fmt.Sprintf("%s\r\n%s", h, temp)
		}
		*header = temp
		message = fmt.Sprintf("GET %s HTTP/1.1\r\n%s\r\n", path, temp)
	} else {
		message = fmt.Sprintf("GET %s HTTP/1.1\r\n\r\n", path)
	}
	if _, err = io.WriteString(conn, message); err != nil {
		return nil, err
	}
	//if _, err = io.WriteString(conn, message); err != nil {
	//	return nil, err
	//}

	req := &http.Request{Method: "GET"}
	resp, err = http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
