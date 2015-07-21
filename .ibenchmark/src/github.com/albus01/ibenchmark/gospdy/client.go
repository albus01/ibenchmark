// Copyright 2013 Jamie Hall. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package spdy

import (
	"errors"
	"github.com/albus01/ibenchmark/gospdy/common"
	"github.com/albus01/ibenchmark/gospdy/spdy2"
	"github.com/albus01/ibenchmark/gospdy/spdy3"
	"net"
	"net/http"
)

// init modifies http.DefaultClient to use a spdy.Transport, enabling
// support for SPDY in functions like http.Get.
func init() {
	http.DefaultClient = NewClient(false)
}

// NewClientConn is used to create a SPDY connection, using the given
// net.Conn for the underlying connection, and the given Receiver to
// receive server pushes.
func NewClientConn(conn net.Conn, push common.Receiver, version, subversion int) (common.Conn, error) {
	if conn == nil {
		return nil, errors.New("Error: Connection initialised with nil net.conn.")
	}

	switch version {
	case 3:
		out := spdy3.NewConn(conn, nil, subversion)
		out.PushReceiver = push
		return out, nil

	case 2:
		out := spdy2.NewConn(conn, nil)
		out.PushReceiver = push
		return out, nil

	default:
		return nil, errors.New("Error: Unrecognised SPDY version.")
	}
}

// NewClient creates an http.Client that supports SPDY.
func NewClient(insecureSkipVerify bool) *http.Client {
	return &http.Client{Transport: NewTransport(insecureSkipVerify)}
}
