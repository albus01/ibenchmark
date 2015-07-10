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
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	headerRegexp = "^([\\w-]+):\\s*(.+)"
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
	concurrency *int    = flag.Int("c", 1, "concurrency:the worker's number,1 default")
	reqNum      *int    = flag.Int("r", 0, "total requests per connection,0 default")
	dur         *int    = flag.Int("t", 0, "timelimit (msec),0 default")
	keepAlive   *bool   = flag.Bool("k", false, "keep the connections every worker established alive,false default")
	cipherSuite *string = flag.String("s", "TLS_RSA_WITH_RC4_128_SHA", "cipher suite,TLS_RSA_WITH_RC4_128_SHA default")
	method      *string = flag.String("m", "GET", "HTTP Method,GET default")
	headers     *string = flag.String("H", "", "request Headers,empty default")
	body        *string = flag.String("B", "", "request Body,empty default")
	out         *bool   = flag.Bool("o", false, "print response body")
	core        *int    = flag.Int("M", 8, "max cores used,8 default")
	SP          *bool   = flag.Bool("S", false, "turn to SPDY")
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
)

func printHelp() {
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
			fmt.Println(err)
			printHelp()
		}
	}()
	flag.Parse()
	if *help {
		printHelp()
	}
	//http https support only
	url, err := gourl.ParseRequestURI(*url)
	if err != nil {
		printHelp()
	}
	proto = url.Scheme
	if proto != "http" && proto != "https" {
		printHelp()
	}
	if h, p, err := net.SplitHostPort(url.Host); err != nil {
		printHelp()
	} else {
		host = h
		port = p
	}
	if path = url.Path; path == "" {
		path = "/"
	}
	if *headers != "" {
		headers := strings.Split(*headers, ";")
		for _, h := range headers {
			match, err := parseHeader(h, headerRegexp)
			if err != nil {
				fmt.Println(err)
				printHelp()
			}
			header.Set(match[1], match[2])
		}
	}
	if host == "" || port == "" || path == "" || proto == "" {
		printHelp()
	}
	ciphers := strings.Split(*cipherSuite, ",")
	for _, c := range ciphers {
		cipherSuites = append(cipherSuites, CipherSuites[c])
	}

	runtime.GOMAXPROCS(*core)

	timeout := time.Duration(*dur) * time.Millisecond
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
	if *keepAlive {
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
		tr = &spdy.Transport{}
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
			//fmt.Println("finish:", i)
			start <- true
			<-end
		}
		finChan <- true
	}

}

//parse headers:'header1:v1;header2:v2'
func parseHeader(in, reg string) (matches []string, err error) {
	re := regexp.MustCompile(reg)
	matches = re.FindStringSubmatch(in)
	if len(matches) < 1 {
		err = errors.New(fmt.Sprintf("Could not parse provided input:%s", err.Error()))
	}
	return
}
