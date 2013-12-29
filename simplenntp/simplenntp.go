// Copyright (c) 2009 The Go Authors. All rights reserved.
// See the LICENSE file.

package simplenntp

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// Connection timeout in seconds
var timeout = time.Duration(20) * time.Second

// A ProtocolError represents responses from an NNTP server
// that seem incorrect for NNTP.
type ProtocolError string
func (p ProtocolError) Error() string {
        return string(p)
}

// An Error represents an error response from an NNTP server.
type Error struct {
        Code uint
        Msg  string
}
func (e Error) Error() string {
        return fmt.Sprintf("%03d %s", e.Code, e.Msg)
}


type Conn struct {
	conn  io.WriteCloser
	r     *bufio.Reader
	close bool
}

func newConn(c net.Conn) (res *Conn, err error) {
	res = &Conn{
		conn: c,
		r:    bufio.NewReaderSize(c, 4096),
	}

	_, err = res.r.ReadString('\n')
	if err != nil {
		return
	}

	return
}

// Dial connects to an NNTP server
func Dial(address string, port int, useTLS bool) (*Conn, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, port), timeout)
	if err != nil {
		return nil, err
	}

	if useTLS {
		// Create and handshake a TLS connection
		tlsConn := tls.Client(conn, nil)
		err = tlsConn.Handshake()
		if err != nil {
			return nil, err
		}

		return newConn(tlsConn)
	} else {
		return newConn(conn)
	}
}

// DialTLS connects to an NNTP server and handles TLS scariness
func DialTLS(address string, port int) (*Conn, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, port), timeout)
	if err != nil {
		return nil, err
	}


	return newConn(conn)
}

// cmd executes an NNTP command:
// It sends the command given by the format and arguments, and then
// reads the response line. If expectCode > 0, the status code on the
// response line must match it. 1 digit expectCodes only check the first
// digit of the status code, etc.
func (c *Conn) cmd(expectCode uint, format string, args ...interface{}) (code uint, line string, err error) {
	if c.close {
		return 0, "", ProtocolError("connection closed")
	}
	// if c.br != nil {
	// 	if err := c.br.discard(); err != nil {
	// 		return 0, "", err
	// 	}
	// 	c.br = nil
	// }
	if _, err := fmt.Fprintf(c.conn, format+"\r\n", args...); err != nil {
		return 0, "", err
	}
	line, err = c.r.ReadString('\n')
	if err != nil {
		return 0, "", err
	}
	line = strings.TrimSpace(line)
	if len(line) < 4 || line[3] != ' ' {
		return 0, "", ProtocolError("short response: " + line)
	}
	i, err := strconv.ParseUint(line[0:3], 10, 0)
	if err != nil {
		return 0, "", ProtocolError("invalid response code: " + line)
	}
	code = uint(i)
	line = line[4:]
	if 1 <= expectCode && expectCode < 10 && code/100 != expectCode ||
		10 <= expectCode && expectCode < 100 && code/10 != expectCode ||
		100 <= expectCode && expectCode < 1000 && code != expectCode {
		err = Error{code, line}
	}
	return
}

// Authenticate logs in to the NNTP server.
// It only sends the password if the server requires one.
func (c *Conn) Authenticate(username, password string) error {
	code, _, err := c.cmd(2, "AUTHINFO USER %s", username)
	if code/100 == 3 {
		_, _, err = c.cmd(2, "AUTHINFO PASS %s", password)
	}
	return err
}

// Post posts an article
func (c *Conn) Post(p []byte) error {
	if _, _, err := c.cmd(3, "POST"); err != nil {
		return err
	}

	start := 0
	end := len(p)
	for {
		n, err := c.conn.Write(p[start:end])
		if err != nil {
			return err
		}

		start += n
		if start == end {
			break
		}
	}

	if _, _, err := c.cmd(240, "."); err != nil {
		return err
	}
	return nil
}

// Quit sends the QUIT command and closes the connection to the server.
func (c *Conn) Quit() error {
	_, _, err := c.cmd(0, "QUIT")
	c.conn.Close()
	c.close = true
	return err
}
