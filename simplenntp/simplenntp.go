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

type TimeData struct {
	Milliseconds int64
	Bytes int
}

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
	tdchan chan *TimeData
	close bool
}

func newConn(c net.Conn, tdchan chan *TimeData) (res *Conn, err error) {
	res = &Conn{
		conn: c,
		r:    bufio.NewReaderSize(c, 4096),
		tdchan: tdchan,
	}

	_, err = res.r.ReadString('\n')
	if err != nil {
		return
	}

	return
}

// Dial connects to an NNTP server
func Dial(address string, port int, useTLS bool, insecureSSL bool, tdchan chan *TimeData) (*Conn, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, port), timeout)
	if err != nil {
		return nil, err
	}

	if useTLS {
		// Create and handshake a TLS connection
		tlsConn := tls.Client(conn, &tls.Config{InsecureSkipVerify: insecureSSL})
		err = tlsConn.Handshake()
		if err != nil {
			return nil, err
		}

		return newConn(tlsConn, tdchan)
	} else {
		return newConn(conn, tdchan)
	}
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
func (c *Conn) Post(p []byte, chunkSize int64) error {
	if _, _, err := c.cmd(3, "POST"); err != nil {
		return err
	}

	plen := int64(len(p))
	start := int64(0)
	end := min(plen, chunkSize)

	for {
		n, err := c.conn.Write(p[start:end])
		if err != nil {
			return err
		}

		// Write a data sent time point to our channel
		c.tdchan <- &TimeData{
			Milliseconds: time.Now().UnixNano() / 1e6,
			Bytes: n,
		}

		// Calculate the next indexes
		start += int64(n)
		end = min(plen, start + chunkSize)
		if start == plen {
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

func min(a, b int64) int64 {
	if (a < b) {
		return a
	}
	return b
}
