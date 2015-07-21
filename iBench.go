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
	"bytes"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/albus01/ibenchmark/bench"
	"github.com/albus01/ibenchmark/gospdy"
	"io"
	"net"
	"net/http"
	gourl "net/url"
	"os"
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
var headers flagHeader
var (
	help        *bool   = flag.Bool("h", false, "show help")
	url         *string = flag.String("u", "https://0.0.0.0:28080/", "server url")
	concurrency *int    = flag.Int("c", 1, "concurrency:the worker's number,1 default")
	reqNum      *int    = flag.Int("r", 1, "total requests per connection,1 default")
	dur         *int    = flag.Int("t", 0, "timelimit (second),0 second default")
	keepAlive   *bool   = flag.Bool("k", false, "keep the connections each worker established alive,false default")
	cipherSuite *string = flag.String("s", "TLS_RSA_WITH_RC4_128_SHA", "cipher suite,TLS_RSA_WITH_RC4_128_SHA default")
	method      *string = flag.String("m", "GET", "HTTP Method,GET default")
	//headers      = flag.Value("H", []string{}, "request Headers,empty default").([]string{})
	body *string = flag.String("B", "", "request Body,empty default")
	out  *bool   = flag.Bool("o", false, "print response body")
	core *int    = flag.Int("M", 8, "max cores used,8 default")
	SP   *bool   = flag.Bool("S", false, "turn to SPDY")
)

var (
	proto        string
	host         string
	port         string
	path         string
	swithHttp    bool            = false
	network      string          = "tcp"
	servers      map[string]bool = make(map[string]bool)
	header       http.Header     = make(http.Header)
	cipherSuites []uint16
	portMap      = map[string]string{"http": "80", "https": "443"}
)

type flagHeader []string

func (f *flagHeader) String() string {
	return fmt.Sprint(headers)
}

func (f *flagHeader) Set(value string) error {
	if headers == nil {
		headers = make(flagHeader, 1)
	} else {
		nheaders := make(flagHeader, len(headers)+1)
		copy(nheaders, headers)
		headers = nheaders
	}
	headers[len(headers)-1] = value
	return nil
}

func printHelp(err interface{}) {
	fmt.Println(err)
	fmt.Println("Usage: iBenchmark [options]")
	flag.PrintDefaults()
	fmt.Printf("\ncihper suite:\n")
	for k := range CipherSuites {
		fmt.Printf("  %s\n", k)
	}
	os.Exit(1)
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			printHelp(err)
		}
	}()
	flag.Var(&headers, "H", "-H \"xxx\" -H \"xxx\" to set muilty headers")
	flag.Parse()
	if *help {
		printHelp(nil)
	}
	//http https support only
	url, err := gourl.ParseRequestURI(*url)
	if err != nil {
		printHelp(err)
	}
	proto = url.Scheme
	if proto != "http" && proto != "https" {
		printHelp(errors.New("only support http or https"))
	}
	if h, p, err := net.SplitHostPort(canonicalAddr(url)); err != nil {
		printHelp(err)
	} else {
		host = h
		port = p
	}
	if path = url.Path; path == "" {
		path = "/"
	}
	if headers != nil {
		for _, h := range headers {
			index := strings.Index(h, ":")
			if index == -1 {
				printHelp(errors.New("Header format error"))
			}
			header.Set(h[:index], h[index+1:])
		}
	}
	if host == "" || port == "" || path == "" || proto == "" {
		printHelp("host port path proto must have value")
	}
	ciphers := strings.Split(*cipherSuite, ",")
	for _, c := range ciphers {
		cipherSuites = append(cipherSuites, CipherSuites[c])
	}

	runtime.GOMAXPROCS(*core)

	timeout := time.Duration(*dur) * time.Second
	finChan := make([]chan bool, *concurrency)

	// number of connections to crypto server cluster
	reporter := new(ibench.Reporter)
	reporter.Concurrency = *concurrency
	reporter.Hostname = host
	reporter.Port = port
	reporter.Path = path

	fmt.Println("ibenchmark start ")
	// start workers
	start := time.Now()
	for i := 0; i < *concurrency; i = i + 1 {
		finChan[i] = make(chan bool)
		go worker(*reqNum, timeout, reporter, finChan[i])
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
	t := float64(reporter.TimeDur) / 1000
	if *keepAlive {
		if t == 0 {
			reporter.RequestPerSecond = 0
		} else {
			reporter.RequestPerSecond = int(float64(reporter.TotalRequest) / t)
		}
		reporter.ConnectionPerSecond = 0
	} else {
		if t == 0 {
			reporter.ConnectionPerSecond = 0
		} else {
			reporter.ConnectionPerSecond = int(float64(reporter.TotalRequest) / t)
		}
		reporter.RequestPerSecond = 0
	}
	var server string
	for key, _ := range servers {
		server = fmt.Sprintf("%s %s", server, key)
	}
	//generate header info
	reporter.Server = server
	for k, v := range header {
		var val string
		for _, v := range v {
			val += v + " "
		}
		reporter.Headers += k + ":" + val + "\r\n"
	}
	time.Sleep(1 * time.Second)
	reporter.Printer()
}
func canonicalAddr(url *gourl.URL) string {
	addr := url.Host
	if !hasPort(addr) {
		return addr + ":" + portMap[url.Scheme]
	}
	return addr
}
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

//and the queries depend on the param dur or requests.if both were setted,depend on dur.See worker func.
//otherwise close the connection immediately when established.
func handle_request(start, done chan bool, client *http.Client, r *ibench.Reporter) {
	for {
		<-start
		var resp *http.Response
		var err error
		var bout bytes.Buffer
		//req, err := http.NewRequest(*method, "https://www.baidu.com", strings.NewReader(*body))
		req, err := http.NewRequest(*method, *url, strings.NewReader(*body))
		if err != nil {
			r.FailedRequest += 1
			done <- true
			continue
		}
		req.Header = header
		if header.Get("Host") != "" {
			//I think this should be a golang http pkg's bug.
			//if I put Host Header in the req.Header,golang pkg can't handle it.
			//So I have to hanlde the Host header in my code.
			req.Host = header.Get("Host")
		}

		r.TotalRequest += 1
		//resp, err = client.Get("https://www.baidu.com")
		resp, err = client.Do(req)
		if err != nil {
			r.FailedRequest += 1
			done <- true
			continue
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
			if *out {
				io.Copy(&bout, resp.Body)
				if bout.String() != "" {
					fmt.Println(bout.String())
				}
			}

			if err := resp.Body.Close(); err != nil {
				r.FailedRequest += 1
			}
		}
		done <- true
	}
}

func request_done(done, end chan bool, r *ibench.Reporter) {
	for {
		start_time := time.Now()
		<-done
		end_time := time.Now()
		elapse := end_time.Sub(start_time).Nanoseconds() / 1000
		r.TimeTaken += elapse
		end <- true
	}
}

//init a go routine,send queries on the transport layer ,the queries number depend on the reqNum or timeout.
//And if both were setted,depends on timeout.
//the finChan notify the main process wether this go routine has finished
func worker(reqNum int, timeout time.Duration, reporter *ibench.Reporter, finChan chan bool) {
	config := tls.Config{
		InsecureSkipVerify:     true,
		SessionTicketsDisabled: true,
		CipherSuites:           cipherSuites,
	}
	var tr http.RoundTripper
	switch {
	case *SP:
		tr = &spdy.Transport{
			TLSClientConfig:   &config,
			DisableKeepAlives: !*keepAlive,
		}
		//tr = spdy.NewTransport(true)
	default:
		tr = &ibench.Transport{
			DisableKeepAlives: !*keepAlive,
			TLSClientConfig:   &config,
		}
	}
	start := make(chan bool, reqNum)
	done := make(chan bool, reqNum)
	end := make(chan bool, reqNum)
	client := &http.Client{Transport: tr}
	end_time := time.After(timeout)
	if *dur != 0 {
		go handle_request(start, done, client, reporter)
		go request_done(done, end, reporter)
		for {
			select {
			case <-end_time:
				finChan <- true
				return
			default:
				start <- true
				<-end
			}
		}

	} else {
		go handle_request(start, done, client, reporter)
		go request_done(done, end, reporter)
		for i := 0; i < reqNum; i++ {
			start <- true
			<-end
		}
		finChan <- true
	}

}
