package netcraze

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"time"
)

var errExpectTimeout = errors.New("expect: timeout waiting for pattern")

type asyncReader struct {
	data chan []byte
	err  chan error
}

func newAsyncReader(r io.Reader) *asyncReader {
	ar := &asyncReader{
		data: make(chan []byte),
		err:  make(chan error, 1),
	}

	go func() {
		buf := make([]byte, 8192)

		for {
			n, readErr := r.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				ar.data <- chunk
			}

			if readErr != nil {
				ar.err <- readErr

				return
			}
		}
	}()

	return ar
}

type console struct {
	reader    *asyncReader
	buf       bytes.Buffer
	closed    bool
	closedErr error
}

func newConsole(r io.Reader) *console {
	return &console{reader: newAsyncReader(r)}
}

func (c *console) expect(re *regexp.Regexp, timeout time.Duration) (string, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		if loc := re.FindIndex(c.buf.Bytes()); loc != nil {
			all := c.buf.Bytes()
			response := string(all[:loc[1]])
			rest := append([]byte(nil), all[loc[1]:]...)
			c.buf.Reset()
			c.buf.Write(rest)

			return response, nil
		}

		if c.closed {
			return c.buf.String(), c.closedErr
		}

		select {
		case chunk := <-c.reader.data:
			c.buf.Write(chunk)
		case err := <-c.reader.err:
			c.closed = true
			c.closedErr = err
		case <-timer.C:
			return c.buf.String(), errExpectTimeout
		}
	}
}

func (c *console) waitClosed(timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		if c.closed {
			return c.closedErr
		}

		select {
		case <-c.reader.data:
		case err := <-c.reader.err:
			c.closed = true
			c.closedErr = err
		case <-timer.C:
			return errExpectTimeout
		}
	}
}
