package memcache

import (
	"bufio"
	"errors"
	"net"
	"strconv"
	"strings"
)

const (
	newline = "\r\n"
)

var (
	ErrFlushAllFailure = errors.New("memcache: Could not flush_all")
	ErrSetFailure      = errors.New("memcache: Could not set")
)

type Connection struct {
	conn net.Conn
	rw   bufio.ReadWriter
}

func Connect(network, address string) (*Connection, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &Connection{
		conn: conn,
		rw: bufio.ReadWriter{
			Reader: bufio.NewReader(conn),
			Writer: bufio.NewWriter(conn),
		},
	}, nil
}
func (conn *Connection) Disconnect() {
	defer conn.conn.Close()
}
func (conn *Connection) FlushAll() (bool, error) {
	err := conn.writeString("flush_all" + newline)
	if err != nil {
		return false, err
	}

	str, err := conn.readLine()
	if !strings.HasPrefix(str, "OK") {
		return false, ErrFlushAllFailure
	}

	return true, nil
}
func (conn *Connection) Get(key string) (ret string, err error) {
	_, err = conn.rw.WriteString("get " + key + newline)
	if err != nil {
		return
	}

	err = conn.rw.Flush()
	if err != nil {
		return
	}

	var str string
	for loop := true; loop; {
		str, err = conn.readLine()
		if err != nil {
			return
		}
		switch {
		case strings.HasPrefix(str, "VALUE"):
			ret, err = conn.readLine()
			if err != nil {
				return
			}
		case strings.HasPrefix(str, "END"):
			loop = false
		case strings.HasPrefix(str, "ERROR"):
			loop = false
		}
	}

	return
}
func (conn *Connection) Set(key, value string) (isSuccess bool, err error) {
	err = conn.writeString("set " + key + " 0 0 " + strconv.Itoa(len(value)) + newline)
	if err != nil {
		return
	}

	err = conn.writeString(value + newline)
	if err != nil {
		return
	}

	str, err := conn.readLine()
	if strings.Contains(str, "ERROR") {
		err = ErrSetFailure
		return
	}

	isSuccess = strings.HasPrefix(str, "STORED")

	return
}
func (conn *Connection) readLine() (string, error) {
	conn.rw.Flush()
	var str string
	for {
		line, isPrefix, err := conn.rw.ReadLine()
		if err != nil {
			return "", err
		}
		str += string(line)
		if !isPrefix {
			break
		}
	}
	return string(str), nil
}
func (conn *Connection) writeString(s string) error {
	srcSize := len(s)
	bufSize := conn.rw.Writer.Size()
	for i := 0; i < srcSize; {
		end := i + bufSize
		if end > len(s) {
			end = len(s)
		}
		conn.rw.Write([]byte(s[i:end]))
		i += bufSize
	}
	return nil
}
