package memcache

import (
	"bufio"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	newline     = "\r\n"
	dialTimeout = 5 * time.Second
)

var (
	ErrFlushAllFailure = errors.New("memcache: Could not flush_all")
	ErrSetFailure      = errors.New("memcache: Could not set")
	ErrDeleteFailure   = errors.New("memcache: Could not delete")
)

type Connection struct {
	conn net.Conn
	rw   bufio.ReadWriter
}

func Connect(network, address string) (*Connection, error) {
	conn, err := net.DialTimeout(network, address, dialTimeout)
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
	conn.conn.Close()
}
func (conn *Connection) FlushAll() (bool, error) {
	err := conn.writeString("flush_all" + newline)
	if err != nil {
		return false, err
	}

	str, err := conn.readLine()
	if err != nil {
		return false, err
	}
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
func (conn *Connection) Set(key, value string) (bool, error) {
	return conn.storeCommand("set", key, value, 0)
}

func (conn *Connection) SetEx(key, value string, exptime int) (bool, error) {
	return conn.storeCommand("set", key, value, exptime)
}

func (conn *Connection) Add(key, value string) (bool, error) {
	return conn.storeCommand("add", key, value, 0)
}

func (conn *Connection) Replace(key, value string) (bool, error) {
	return conn.storeCommand("replace", key, value, 0)
}

func (conn *Connection) Delete(key string) (bool, error) {
	err := conn.writeString("delete " + key + newline)
	if err != nil {
		return false, err
	}
	str, err := conn.readLine()
	if err != nil {
		return false, err
	}
	if strings.Contains(str, "ERROR") {
		return false, ErrDeleteFailure
	}
	return strings.HasPrefix(str, "DELETED"), nil
}

func (conn *Connection) Increment(key string, delta uint64) (uint64, error) {
	return conn.incrDecrCommand("incr", key, delta)
}

func (conn *Connection) Decrement(key string, delta uint64) (uint64, error) {
	return conn.incrDecrCommand("decr", key, delta)
}

func (conn *Connection) storeCommand(command, key, value string, exptime int) (bool, error) {
	err := conn.writeString(command + " " + key + " 0 " + strconv.Itoa(exptime) + " " + strconv.Itoa(len(value)) + newline)
	if err != nil {
		return false, err
	}
	err = conn.writeString(value + newline)
	if err != nil {
		return false, err
	}
	str, err := conn.readLine()
	if err != nil {
		return false, err
	}
	if strings.Contains(str, "ERROR") {
		return false, ErrSetFailure
	}
	return strings.HasPrefix(str, "STORED"), nil
}

func (conn *Connection) incrDecrCommand(command, key string, delta uint64) (uint64, error) {
	err := conn.writeString(command + " " + key + " " + strconv.FormatUint(delta, 10) + newline)
	if err != nil {
		return 0, err
	}
	str, err := conn.readLine()
	if err != nil {
		return 0, err
	}
	if strings.Contains(str, "ERROR") || strings.HasPrefix(str, "NOT_FOUND") {
		return 0, errors.New("memcache: " + command + " failed: " + str)
	}
	return strconv.ParseUint(strings.TrimSpace(str), 10, 64)
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
		_, err := conn.rw.Write([]byte(s[i:end]))
		if err != nil {
			return err
		}
		i += bufSize
	}
	return nil
}
