package memcachemanager

import (
	"log"
	"testing"

	"github.com/rssh-jp/go-memcache/memcache"
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

func TestConnection(t *testing.T) {
	conf := Config{
		ParallelNum: 2,
		ConnectionList: []MemcacheConnection{
			{"mem1", "tcp", "localhost", "11211"},
		},
	}
	err := Initialize(conf)
	if err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	conf := Config{
		ParallelNum: 2,
		ConnectionList: []MemcacheConnection{
			{"mem1", "tcp", "localhost", "11211"},
		},
	}
	err := Initialize(conf)
	if err != nil {
		t.Error(err)
	}

	for _, item := range []struct {
		key, value string
		isSuccess  bool
	}{
		{"test", "test", true},
		{"test1", "", false},
	} {
		res, err := Get(item.key)
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
	conf := Config{
		ParallelNum: 2,
		ConnectionList: []MemcacheConnection{
			{"mem1", "tcp", "localhost", "11211"},
			{"mem2", "tcp", "localhost", "11211"},
		},
	}
	err := Initialize(conf)
	if err != nil {
		t.Error(err)
	}

	for _, item := range []struct {
		key, value string
	}{
		{"test1", "test1"},
		{"test2", "test2"},
	} {
		err := Set(item.key, item.value)
		if err != nil {
			t.Error(err)
			return
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
		err := Set(string(key), value)
		if err != nil {
			t.Error("Could not set memcache long key.", err)
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
		err := Set(key, string(value))
		if err != nil {
			t.Error("Could not set memcache long value.", err)
			return
		}
	}
}

func TestSetEx(t *testing.T) {
	conf := Config{
		ParallelNum: 2,
		ConnectionList: []MemcacheConnection{
			{"mem1", "tcp", "localhost", "11211"},
		},
	}
	err := Initialize(conf)
	if err != nil {
		t.Error(err)
	}

	err = SetEx("setex_mgr_test", "hello", 60)
	if err != nil {
		t.Error(err)
		return
	}

	val, err := Get("setex_mgr_test")
	if err != nil {
		t.Error(err)
		return
	}
	if val != "hello" {
		t.Errorf("SetEx: got %s, want hello", val)
	}
}

func TestDelete(t *testing.T) {
	conf := Config{
		ParallelNum: 2,
		ConnectionList: []MemcacheConnection{
			{"mem1", "tcp", "localhost", "11211"},
		},
	}
	err := Initialize(conf)
	if err != nil {
		t.Error(err)
	}

	err = Set("del_mgr_test", "value")
	if err != nil {
		t.Error(err)
		return
	}

	err = Delete("del_mgr_test")
	if err != nil {
		t.Error(err)
		return
	}

	val, err := Get("del_mgr_test")
	if err != nil {
		t.Error(err)
		return
	}
	if val != "" {
		t.Errorf("Delete: expected empty after delete, got %s", val)
	}
}
