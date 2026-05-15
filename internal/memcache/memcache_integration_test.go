//go:build integration

package memcache_test

import (
	"log"
	"testing"

	"github.com/rssh-jp/go-memcache/internal/memcache"
)

func initMemcache() {
	conn, err := memcache.Connect("tcp", "localhost:11211")
	if err != nil {
		log.Fatal("Could not create connection.")
	}

	ok, err := conn.FlushAll()
	if err != nil {
		log.Fatal("Could not flush_all memcache.", err)
	}
	if !ok {
		log.Fatal("Could not flush_all memcache.")
	}

	ok, err = conn.Set("test", "test")
	log.Println(ok, err)
	if err != nil {
		log.Fatal("Could not set memcache.", err)
	}
	if !ok {
		log.Fatal("Could not set memcache.")
	}
}

func TestMain(m *testing.M) {
	initMemcache()
	m.Run()
}

func TestConnect(t *testing.T) {
	for _, item := range []struct {
		network, address string
		isSuccess        bool
	}{
		{"tcp", "localhost:11211", true},
		{"tcp", "localhost:11212", false},
	} {
		_, err := memcache.Connect(item.network, item.address)
		if item.isSuccess {
			if err != nil {
				t.Error(err)
			}
		} else {
			if err == nil {
				t.Error("Invalid connection success.")
			}
		}
	}
}

func TestGet(t *testing.T) {
	conn, err := memcache.Connect("tcp", "localhost:11211")
	if err != nil {
		t.Error(err)
		return
	}

	for _, item := range []struct {
		key, value string
		isSuccess  bool
	}{
		{"test", "test", true},
		{"test1", "", false},
	} {
		res, err := conn.Get(item.key)
		if err != nil {
			t.Error(err)
			return
		}
		if res != item.value {
			if item.isSuccess {
				t.Errorf("got %s, want %s", res, item.value)
			} else {
				t.Errorf(" ng pattern: got %s, want %s", res, item.value)
			}
		}
	}
}

func TestSet(t *testing.T) {
	conn, err := memcache.Connect("tcp", "localhost:11211")
	if err != nil {
		t.Error(err)
		return
	}

	for _, item := range []struct {
		key, value string
		isSuccess  bool
	}{
		{"test1", "test1", true},
		{"test2", "test2", true},
	} {
		ok, err := conn.Set(item.key, item.value)
		if err != nil {
			t.Error(err)
			return
		}
		if ok != item.isSuccess {
			t.Error("Could not set memcache.")
		}
	}

	// キーの最大値テスト
	{
		const length = 250
		key := make([]byte, 0, length)
		for i := 0; i < length; i++ {
			key = append(key, 49)
		}
		value := "test"
		ok, err := conn.Set(string(key), value)
		if err != nil {
			t.Error("Could not set memcache long key.", err)
			return
		}
		if !ok {
			t.Error("Could not set memcache long key.")
			return
		}
	}

	// でかいサイズの値のテスト
	{
		const length = 1 * 1000 * 1000
		key := "test"
		value := make([]byte, 0, length)
		for i := 0; i < length; i++ {
			value = append(value, 49)
		}
		ok, err := conn.Set(key, string(value))
		if err != nil {
			t.Error("Could not set memcache long value.", err)
			return
		}
		if !ok {
			t.Error("Could not set memcache long value.")
			return
		}
	}
}

func TestSetEx(t *testing.T) {
	conn, err := memcache.Connect("tcp", "localhost:11211")
	if err != nil {
		t.Error(err)
		return
	}

	ok, err := conn.SetEx("setex_test", "hello", 60)
	if err != nil {
		t.Error(err)
		return
	}
	if !ok {
		t.Error("SetEx: expected STORED")
		return
	}

	val, err := conn.Get("setex_test")
	if err != nil {
		t.Error(err)
		return
	}
	if val != "hello" {
		t.Errorf("SetEx: got %s, want hello", val)
	}
}

func TestAdd(t *testing.T) {
	conn, err := memcache.Connect("tcp", "localhost:11211")
	if err != nil {
		t.Error(err)
		return
	}

	// 存在しないキーへの Add は成功する
	conn.Delete("add_test")
	ok, err := conn.Add("add_test", "first")
	if err != nil {
		t.Error(err)
		return
	}
	if !ok {
		t.Error("Add: expected STORED for new key")
		return
	}

	// 同じキーへの再 Add は NOT_STORED（false, nil）
	ok, err = conn.Add("add_test", "second")
	if err != nil {
		t.Error(err)
		return
	}
	if ok {
		t.Error("Add: expected NOT_STORED for existing key")
	}
}

func TestReplace(t *testing.T) {
	conn, err := memcache.Connect("tcp", "localhost:11211")
	if err != nil {
		t.Error(err)
		return
	}

	// 存在するキーへの Replace は成功する
	conn.Set("replace_test", "original")
	ok, err := conn.Replace("replace_test", "replaced")
	if err != nil {
		t.Error(err)
		return
	}
	if !ok {
		t.Error("Replace: expected STORED for existing key")
		return
	}

	val, err := conn.Get("replace_test")
	if err != nil {
		t.Error(err)
		return
	}
	if val != "replaced" {
		t.Errorf("Replace: got %s, want replaced", val)
	}

	// 存在しないキーへの Replace は NOT_STORED（false, nil）
	conn.Delete("replace_nokey")
	ok, err = conn.Replace("replace_nokey", "value")
	if err != nil {
		t.Error(err)
		return
	}
	if ok {
		t.Error("Replace: expected NOT_STORED for non-existent key")
	}
}

func TestDelete(t *testing.T) {
	conn, err := memcache.Connect("tcp", "localhost:11211")
	if err != nil {
		t.Error(err)
		return
	}

	conn.Set("delete_test", "value")

	// 存在するキーの削除は true
	ok, err := conn.Delete("delete_test")
	if err != nil {
		t.Error(err)
		return
	}
	if !ok {
		t.Error("Delete: expected DELETED")
		return
	}

	// 削除後は Get で空文字
	val, err := conn.Get("delete_test")
	if err != nil {
		t.Error(err)
		return
	}
	if val != "" {
		t.Errorf("Delete: expected empty after delete, got %s", val)
	}

	// 存在しないキーの削除は false, nil
	ok, err = conn.Delete("nonexistent_key_xyz")
	if err != nil {
		t.Error(err)
		return
	}
	if ok {
		t.Error("Delete: expected NOT_FOUND for non-existent key")
	}
}

func TestIncrDecr(t *testing.T) {
	conn, err := memcache.Connect("tcp", "localhost:11211")
	if err != nil {
		t.Error(err)
		return
	}

	conn.Set("counter", "10")

	n, err := conn.Increment("counter", 5)
	if err != nil {
		t.Error(err)
		return
	}
	if n != 15 {
		t.Errorf("Increment: got %d, want 15", n)
	}

	n, err = conn.Decrement("counter", 3)
	if err != nil {
		t.Error(err)
		return
	}
	if n != 12 {
		t.Errorf("Decrement: got %d, want 12", n)
	}

	// 存在しないキーへの Increment はエラー
	conn.Delete("no_counter")
	_, err = conn.Increment("no_counter", 1)
	if err == nil {
		t.Error("Increment: expected error for non-existent key")
	}
}
