package memcache_test

import (
	"bufio"
	"fmt"
	"net"
	"testing"

	"github.com/rssh-jp/go-memcache/internal/memcache"
)

// newMockConn は net.Pipe() を使ってモックサーバー付きの Connection を返す。
// モックサーバーは cmdLines 行のリクエストを読み捨てた後、response を返す。
func newMockConn(t *testing.T, cmdLines int, response string) *memcache.Connection {
	t.Helper()
	client, server := net.Pipe()
	t.Cleanup(func() {
		client.Close()
		server.Close()
	})
	conn := memcache.NewConnectionFromConn(client)
	go func() {
		br := bufio.NewReader(server)
		for i := 0; i < cmdLines; i++ {
			if _, _, err := br.ReadLine(); err != nil {
				return
			}
		}
		fmt.Fprint(server, response)
	}()
	return conn
}

func TestGet_Hit(t *testing.T) {
	conn := newMockConn(t, 1, "VALUE mykey 0 5\r\nhello\r\nEND\r\n")
	val, err := conn.Get("mykey")
	if err != nil {
		t.Fatal(err)
	}
	if val != "hello" {
		t.Errorf("got %q, want %q", val, "hello")
	}
}

func TestGet_Miss(t *testing.T) {
	conn := newMockConn(t, 1, "END\r\n")
	val, err := conn.Get("nokey")
	if err != nil {
		t.Fatal(err)
	}
	if val != "" {
		t.Errorf("got %q, want empty string", val)
	}
}

func TestSet_Stored(t *testing.T) {
	conn := newMockConn(t, 2, "STORED\r\n")
	ok, err := conn.Set("key", "value")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected STORED (true)")
	}
}

func TestSet_NotStored(t *testing.T) {
	conn := newMockConn(t, 2, "NOT_STORED\r\n")
	ok, err := conn.Set("key", "value")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected NOT_STORED (false)")
	}
}

func TestSetEx_Stored(t *testing.T) {
	conn := newMockConn(t, 2, "STORED\r\n")
	ok, err := conn.SetEx("key", "value", 60)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected STORED (true)")
	}
}

func TestAdd_Stored(t *testing.T) {
	conn := newMockConn(t, 2, "STORED\r\n")
	ok, err := conn.Add("key", "value")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected STORED (true)")
	}
}

func TestAdd_NotStored(t *testing.T) {
	conn := newMockConn(t, 2, "NOT_STORED\r\n")
	ok, err := conn.Add("key", "value")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected NOT_STORED (false)")
	}
}

func TestReplace_Stored(t *testing.T) {
	conn := newMockConn(t, 2, "STORED\r\n")
	ok, err := conn.Replace("key", "value")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected STORED (true)")
	}
}

func TestReplace_NotStored(t *testing.T) {
	conn := newMockConn(t, 2, "NOT_STORED\r\n")
	ok, err := conn.Replace("nokey", "value")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected NOT_STORED (false)")
	}
}

func TestDelete_Deleted(t *testing.T) {
	conn := newMockConn(t, 1, "DELETED\r\n")
	ok, err := conn.Delete("key")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected DELETED (true)")
	}
}

func TestDelete_NotFound(t *testing.T) {
	conn := newMockConn(t, 1, "NOT_FOUND\r\n")
	ok, err := conn.Delete("nokey")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected NOT_FOUND (false)")
	}
}

func TestIncrement(t *testing.T) {
	conn := newMockConn(t, 1, "15\r\n")
	n, err := conn.Increment("counter", 5)
	if err != nil {
		t.Fatal(err)
	}
	if n != 15 {
		t.Errorf("got %d, want 15", n)
	}
}

func TestDecrement(t *testing.T) {
	conn := newMockConn(t, 1, "7\r\n")
	n, err := conn.Decrement("counter", 3)
	if err != nil {
		t.Fatal(err)
	}
	if n != 7 {
		t.Errorf("got %d, want 7", n)
	}
}

func TestFlushAll(t *testing.T) {
	conn := newMockConn(t, 1, "OK\r\n")
	ok, err := conn.FlushAll()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected OK (true)")
	}
}
