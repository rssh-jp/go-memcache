package memcache

import (
	"log"
	"testing"
)

func initMemcache() {
	conn, err := Connect("tcp", "localhost:11211")
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
		_, err := Connect(item.network, item.address)
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
	conn, err := Connect("tcp", "localhost:11211")
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
	conn, err := Connect("tcp", "localhost:11211")
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
