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
package ibench

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

//this transport make you more freedom to control the connection that you can decide the connection should keep-alive or not.Also it's more sample than the standard.
type Transport struct {
	Dial              func(net, addr string) (c net.Conn, err error)
	TLSClientConfig   *tls.Config
	DisableKeepAlives bool
	Conn              net.Conn
}

var (
	portMap = map[string]string{"https": "443", "http": "80"}
)

func (t *Transport) dial(network, addr string) (c net.Conn, err error) {
	if t.Dial != nil {
		return t.Dial(network, addr)
	}
	return net.Dial(network, addr)
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if req.URL == nil {
		return nil, errors.New("http: nil Request.URL")
	}
	if req.Header == nil {
		return nil, errors.New("http: nil Request.Header")
	}
	if req.URL.Host == "" {
		return nil, errors.New("http: no Host in request URL")
	}
	cm := t.connectMethodForRequest(req)
	conn, err := t.getConn(cm)
	if err != nil {
		return nil, err
	}
	if err := req.Write(*conn); err != nil {
		return nil, err
	}
	resp, err = http.ReadResponse(bufio.NewReader(*conn), req)
	if err != nil {
		return nil, err
	}
	return resp, nil

}

func (t *Transport) connectMethodForRequest(treq *http.Request) (cm connectMethod) {
	cm.targetScheme = treq.URL.Scheme
	cm.targetAddr = canonicalAddr(treq.URL)
	return cm
}

func (t *Transport) getConn(cm connectMethod) (*net.Conn, error) {
	if !t.DisableKeepAlives {
		if t.Conn == nil {
			conn, err := t.dialConn(cm)
			if err != nil {
				return nil, err
			}
			t.Conn = *conn
			return conn, nil
		} else {
			return &t.Conn, nil
		}
	}
	return t.dialConn(cm)
}

func (t *Transport) dialConn(cm connectMethod) (*net.Conn, error) {
	var conn net.Conn
	var err error
	if cm.targetScheme == "https" {
		if t.TLSClientConfig == nil {
			t.TLSClientConfig = &tls.Config{
				InsecureSkipVerify:     true,
				SessionTicketsDisabled: true,
			}
		}
		conn, err = tls.Dial("tcp", cm.targetAddr, t.TLSClientConfig)
		if err != nil {
			return nil, err
		}
		return &conn, nil
	} else if cm.targetScheme == "http" {
		conn, err = t.dial("tcp", cm.addr())
		if err != nil {
			return nil, err
		}
		return &conn, err
	}
	return nil, errors.New(fmt.Sprintf("Do not support the schema:%s", cm.targetAddr))
}

//func (t *Transport) dialConn(cm connectMethod) (*net.Conn, error) {
//	var conn net.Conn
//	if cm.targetScheme == "https" {
//		if t.TLSClientConfig == nil {
//			t.TLSClientConfig = &tls.Config{}
//		}
//		c, err := tls.Dial("tcp", cm.targetAddr, t.TLSClientConfig)
//		if err != nil {
//			return nil, err
//		}
//		conn = *c
//		return &conn, nil
//	} else if cm.targetScheme == "http" {
//		c, err := t.dial("tcp", cm.addr())
//		if err != nil {
//			return nil, err
//		}
//		conn = c.(net.Conn)
//		return &conn, err
//	}
//	return nil, errors.New(fmt.Sprintf("Do not support the schema:%s", cm.targetAddr))
//}

// canonicalAddr returns url.Host but always with a ":port" suffix
func canonicalAddr(url *url.URL) string {
	addr := url.Host
	if !hasPort(addr) {
		return addr + ":" + portMap[url.Scheme]
	}
	return addr
}

func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

type connectMethod struct {
	proxyURL     *url.URL // nil for no proxy, else full proxy URL
	targetScheme string   // "http" or "https"
	targetAddr   string   // Not used if proxy + http targetScheme (4th example in table)
}

// addr returns the first hop "host:port" to which we need to TCP connect.
func (cm *connectMethod) addr() string {
	if cm.proxyURL != nil {
		return canonicalAddr(cm.proxyURL)
	}
	return cm.targetAddr
}
