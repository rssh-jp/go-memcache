package memcachemanager

import (
	"errors"
	"log"
	"time"

	"memcache/memcache"
)

const (
	shortTimeOut = time.Nanosecond
)

var (
	connections = make(map[string]chan *memcache.Connection)
	timeOut     = time.Second * 5
)

type Config struct {
	ParallelNum    int
	ConnectionList []MemcacheConnection
}
type MemcacheConnection struct {
	Name, Network, Host, Port string
}

func Initialize(conf Config) (err error) {
	for _, info := range conf.ConnectionList {
		connections[info.Name] = make(chan *memcache.Connection, conf.ParallelNum)
		for i := 0; i < conf.ParallelNum; i++ {
			conn, err := memcache.Connect(info.Network, info.Host+":"+info.Port)
			if err != nil {
				return err
			}
			connections[info.Name] <- conn
		}
	}
	return
}

func Get(key string) (ret string, err error) {
	conn, name, err := searchConnection()
	if err != nil {
		return
	}

	defer func() {
		if conn == nil {
			return
		}
		connections[name] <- conn
	}()

	ret, err = conn.Get(key)
	if err != nil {
		return
	}

	return
}

func Set(name, key, value string) (err error) {
	conns, ok := connections[name]
	if !ok {
		err = errors.New("Could not found memcache connection. name=" + name)
		log.Println(err)
		return
	}

	var conn *memcache.Connection
	select {
	case conn = <-conns:
	case <-time.After(timeOut):
		err = errors.New("Timeout")
		log.Println(err)
		return
	}

	defer func() {
		if conn == nil {
			return
		}
		connections[name] <- conn
	}()

	isSuccess, err := conn.Set(key, value)
	if err != nil {
		return
	}
	if !isSuccess {
		err = errors.New("Could not success memcache set.")
		return
	}

	return
}

// コネクション探索
func searchConnection() (conn *memcache.Connection, name string, err error) {
	// 最速で接続できるところを探す
	conn, name = searchConnectionTimeout(shortTimeOut)

	// 接続先が見つからなかったら通常のタイムアウト間隔で接続できるところを探す
	if conn == nil {
		conn, name = searchConnectionTimeout(timeOut)
		// 接続先が見つからなかったらタイムアウト
		if conn == nil {
			log.Println("Time out memcached connection.")
			err = errors.New("Timeout")
			return
		}
	}
	return
}

// タイムアウト指定でコネクションを探す
func searchConnectionTimeout(d time.Duration) (conn *memcache.Connection, name string) {
	for n, conns := range connections {
		select {
		case conn = <-conns:
			name = n
			return
		case <-time.After(d):
		}
	}
	return
}
