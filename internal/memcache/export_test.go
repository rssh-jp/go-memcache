package memcache

import (
	"bufio"
	"net"
)

// NewConnectionFromConn は既存の net.Conn から Connection を生成する。
// テストで net.Pipe() を使ったモック接続を作るためにのみ使用する。
func NewConnectionFromConn(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
		rw: bufio.ReadWriter{
			Reader: bufio.NewReader(conn),
			Writer: bufio.NewWriter(conn),
		},
	}
}
