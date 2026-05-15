# go-memcache

Memcached へのコネクション管理とデータ操作を提供する Go ライブラリ。

## パッケージ構成

```
go-memcache/
├── memcache/
│   ├── memcache.go       # Memcached 生接続ラッパー（Connection 構造体）
│   └── memcache_test.go
├── memcachemanager.go    # コネクションプール管理（Initialize / Get / Set / SetEx / Delete）
├── memcachemanager_test.go
├── go.mod
├── Makefile
└── README.md
```

- **`memcache`** パッケージ: `net.Conn` を薄くラップし、Memcached テキストプロトコルの主要コマンド（get / set / add / replace / delete / incr / decr / flush_all）を実装する低レイヤライブラリ。
- **`memcachemanager`** パッケージ: `memcache.Connection` をコネクションプールとして管理し、複数の名前付き接続先への並列アクセスを提供する高レイヤライブラリ。

## インストール

```bash
go get github.com/rssh-jp/go-memcache
```

## クイックスタート

### コネクション直接利用（`memcache` パッケージ）

```go
import "github.com/rssh-jp/go-memcache/memcache"

conn, err := memcache.Connect("tcp", "localhost:11211")
if err != nil {
    // handle error
}
defer conn.Disconnect()

// 値をセット（有効期限なし）
ok, err := conn.Set("key", "value")

// 有効期限付きでセット（60秒）
ok, err = conn.SetEx("key", "value", 60)

// キーが存在しない場合のみセット
ok, err = conn.Add("newkey", "value")

// キーが存在する場合のみ更新
ok, err = conn.Replace("key", "updated")

// キーを削除
ok, err = conn.Delete("key")

// 数値のインクリメント / デクリメント
conn.Set("counter", "10")
n, err := conn.Increment("counter", 5) // → 15
n, err  = conn.Decrement("counter", 3) // → 12

// 値を取得
val, err := conn.Get("key")

// 全キーを削除
ok, err = conn.FlushAll()
```

### コネクションプール利用（`memcachemanager` パッケージ）

```go
import memcachemanager "github.com/rssh-jp/go-memcache"

conf := memcachemanager.Config{
    ParallelNum: 3,
    ConnectionList: []memcachemanager.MemcacheConnection{
        {Name: "mem1", Network: "tcp", Host: "localhost", Port: "11211"},
        {Name: "mem2", Network: "tcp", Host: "cache2", Port: "11211"},
    },
}
if err := memcachemanager.Initialize(conf); err != nil {
    // handle error
}

// 最速のコネクションで取得
val, err := memcachemanager.Get("key")

// 全コネクションに書き込み
err = memcachemanager.Set("key", "value")

// 有効期限付きでセット（60秒）
err = memcachemanager.SetEx("key", "value", 60)

// キーを削除
err = memcachemanager.Delete("key")
```

## API リファレンス

### `memcache` パッケージ

| 関数 / メソッド | シグネチャ | 説明 |
|---|---|---|
| `Connect` | `(network, address string) (*Connection, error)` | Memcached に接続し `Connection` を返す |
| `(*Connection).Disconnect` | `()` | 接続を閉じる |
| `(*Connection).FlushAll` | `() (bool, error)` | 全キーを削除する（`flush_all`） |
| `(*Connection).Get` | `(key string) (string, error)` | キーに対応する値を取得する |
| `(*Connection).Set` | `(key, value string) (bool, error)` | キーと値を保存する（有効期限なし） |
| `(*Connection).SetEx` | `(key, value string, exptime int) (bool, error)` | 有効期限（秒）付きでキーと値を保存する |
| `(*Connection).Add` | `(key, value string) (bool, error)` | キーが存在しない場合のみ保存する |
| `(*Connection).Replace` | `(key, value string) (bool, error)` | キーが存在する場合のみ更新する |
| `(*Connection).Delete` | `(key string) (bool, error)` | キーを削除する（存在しない場合は `false, nil`） |
| `(*Connection).Increment` | `(key string, delta uint64) (uint64, error)` | 数値を加算し新しい値を返す |
| `(*Connection).Decrement` | `(key string, delta uint64) (uint64, error)` | 数値を減算し新しい値を返す |

### `memcachemanager` パッケージ

| 関数 | シグネチャ | 説明 |
|---|---|---|
| `Initialize` | `(conf Config) error` | コネクションプールを初期化する |
| `Get` | `(key string) (string, error)` | 最速のコネクションからキーを取得する |
| `Set` | `(key, value string) error` | 全コネクションにキーと値を書き込む |
| `SetEx` | `(key, value string, exptime int) error` | 有効期限（秒）付きで全コネクションに書き込む |
| `Delete` | `(key string) error` | 全コネクションからキーを削除する |

#### `Config` 構造体

| フィールド | 型 | 説明 |
|---|---|---|
| `ParallelNum` | `int` | コネクションプールサイズ（接続ごとの並列数） |
| `ConnectionList` | `[]MemcacheConnection` | 接続先定義リスト |

#### `MemcacheConnection` 構造体

| フィールド | 型 | 説明 |
|---|---|---|
| `Name` | `string` | 接続先を識別する名前 |
| `Network` | `string` | ネットワーク種別（例: `"tcp"`） |
| `Host` | `string` | ホスト名または IP アドレス |
| `Port` | `string` | ポート番号（例: `"11211"`） |

## テスト・検証コマンド

テスト実行には Memcached が `localhost:11211` で起動している必要があります。

```bash
# Memcached を起動（Docker を使う場合）
docker run -d -p 11211:11211 memcached
```

| コマンド | 説明 |
|---|---|
| `make build` | ライブラリをビルドする |
| `make test` | ユニットテストを実行する |
| `make test-verbose` | 詳細ログ付きでテストを実行する |
| `make vet` | 静的解析を実行する |
| `make tidy` | `go.mod` / `go.sum` を整理する |

## 動作要件

- Go 1.26 以上
- Memcached（テキストプロトコル対応）

## バージョン

0.6


## インストール

```bash
go get github.com/rssh-jp/go-memcache
```

## クイックスタート

### コネクション直接利用（`memcache` パッケージ）

```go
import "github.com/rssh-jp/go-memcache/memcache"

conn, err := memcache.Connect("tcp", "localhost:11211")
if err != nil {
    // handle error
}
defer conn.Disconnect()

// 値をセット（有効期限なし）
ok, err := conn.Set("key", "value")

// 有効期限付きでセット（60秒）
ok, err = conn.SetEx("key", "value", 60)

// キーが存在しない場合のみセット
ok, err = conn.Add("newkey", "value")

// キーが存在する場合のみ更新
ok, err = conn.Replace("key", "updated")

// キーを削除
ok, err = conn.Delete("key")

// 数値のインクリメント / デクリメント
conn.Set("counter", "10")
n, err := conn.Increment("counter", 5) // → 15
n, err  = conn.Decrement("counter", 3) // → 12

// 値を取得
val, err := conn.Get("key")

// 全キーを削除
ok, err = conn.FlushAll()
```

### コネクションプール利用（`memcachemanager` パッケージ）

```go
import memcachemanager "github.com/rssh-jp/go-memcache"

conf := memcachemanager.Config{
    ParallelNum: 3,
    ConnectionList: []memcachemanager.MemcacheConnection{
        {Name: "mem1", Network: "tcp", Host: "localhost", Port: "11211"},
        {Name: "mem2", Network: "tcp", Host: "cache2", Port: "11211"},
    },
}
if err := memcachemanager.Initialize(conf); err != nil {
    // handle error
}

// 最速のコネクションで取得
val, err := memcachemanager.Get("key")

// 全コネクションに書き込み
err = memcachemanager.Set("key", "value")

// 有効期限付きでセット（60秒）
err = memcachemanager.SetEx("key", "value", 60)

// キーを削除
err = memcachemanager.Delete("key")
```

## API リファレンス

### `memcache` パッケージ

| 関数 / メソッド | シグネチャ | 説明 |
|---|---|---|
| `Connect` | `(network, address string) (*Connection, error)` | Memcached に接続し `Connection` を返す |
| `(*Connection).Disconnect` | `()` | 接続を閉じる |
| `(*Connection).FlushAll` | `() (bool, error)` | 全キーを削除する（`flush_all`） |
| `(*Connection).Get` | `(key string) (string, error)` | キーに対応する値を取得する |
| `(*Connection).Set` | `(key, value string) (bool, error)` | キーと値を保存する（有効期限なし） |
| `(*Connection).SetEx` | `(key, value string, exptime int) (bool, error)` | 有効期限（秒）付きでキーと値を保存する |
| `(*Connection).Add` | `(key, value string) (bool, error)` | キーが存在しない場合のみ保存する |
| `(*Connection).Replace` | `(key, value string) (bool, error)` | キーが存在する場合のみ更新する |
| `(*Connection).Delete` | `(key string) (bool, error)` | キーを削除する（存在しない場合は `false, nil`） |
| `(*Connection).Increment` | `(key string, delta uint64) (uint64, error)` | 数値を加算し新しい値を返す |
| `(*Connection).Decrement` | `(key string, delta uint64) (uint64, error)` | 数値を減算し新しい値を返す |

### `memcachemanager` パッケージ

| 関数 | シグネチャ | 説明 |
|---|---|---|
| `Initialize` | `(conf Config) error` | コネクションプールを初期化する |
| `Get` | `(key string) (string, error)` | 最速のコネクションからキーを取得する |
| `Set` | `(key, value string) error` | 全コネクションにキーと値を書き込む |
| `SetEx` | `(key, value string, exptime int) error` | 有効期限（秒）付きで全コネクションに書き込む |
| `Delete` | `(key string) error` | 全コネクションからキーを削除する |

#### `Config` 構造体

| フィールド | 型 | 説明 |
|---|---|---|
| `ParallelNum` | `int` | コネクションプールサイズ（接続ごとの並列数） |
| `ConnectionList` | `[]MemcacheConnection` | 接続先定義リスト |

#### `MemcacheConnection` 構造体

| フィールド | 型 | 説明 |
|---|---|---|
| `Name` | `string` | 接続先を識別する名前 |
| `Network` | `string` | ネットワーク種別（例: `"tcp"`） |
| `Host` | `string` | ホスト名または IP アドレス |
| `Port` | `string` | ポート番号（例: `"11211"`） |

## 動作要件

- Go 1.26 以上
- Memcached（テキストプロトコル対応）

## バージョン

0.5
