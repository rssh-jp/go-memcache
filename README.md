# go-memcache

Memcached へのコネクション管理とデータ操作（Get/Set）を提供する Go ライブラリ。

## パッケージ構成

```
go-memcache/
├── memcache/
│   ├── memcache.go       # Memcached 生接続ラッパー（Connection 構造体）
│   └── memcache_test.go
├── memcachemanager.go    # コネクションプール管理（Initialize / Get / Set）
├── memcachemanager_test.go
├── go.mod
└── README.md
```

- **`memcache`** パッケージ: `net.Conn` を薄くラップし、Memcached テキストプロトコルの `flush_all` / `get` / `set` コマンドを実装する。
- **`memcachemanager`** パッケージ: `memcache.Connection` をコネクションプールとして管理し、複数の名前付き接続先への並列アクセスを提供する。

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

// 値をセット
ok, err := conn.Set("key", "value")

// 値を取得
val, err := conn.Get("key")

// 全キーを削除
ok, err := conn.FlushAll()
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

// 名前を指定してセット
err = memcachemanager.Set("mem1", "key", "value")
```

## API リファレンス

### `memcache` パッケージ

| 関数 / メソッド | シグネチャ | 説明 |
|---|---|---|
| `Connect` | `(network, address string) (*Connection, error)` | Memcached に接続し `Connection` を返す |
| `(*Connection).Disconnect` | `()` | 接続を閉じる |
| `(*Connection).FlushAll` | `() (bool, error)` | 全キーを削除する（`flush_all`） |
| `(*Connection).Get` | `(key string) (string, error)` | キーに対応する値を取得する |
| `(*Connection).Set` | `(key, value string) (bool, error)` | キーと値を保存する |

### `memcachemanager` パッケージ

| 関数 | シグネチャ | 説明 |
|---|---|---|
| `Initialize` | `(conf Config) error` | コネクションプールを初期化する |
| `Get` | `(key string) (string, error)` | 最速のコネクションからキーを取得する |
| `Set` | `(name, key, value string) error` | 指定した名前のコネクションに値をセットする |

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

## テスト

テスト実行には Memcached が `localhost:11211` で起動している必要があります。

```bash
# Memcached を起動（Docker を使う場合）
docker run -d -p 11211:11211 memcached

# テスト実行
go test ./...

# ビルド確認
go build ./...

# 静的解析
go vet ./...
```

## 動作要件

- Go 1.26 以上
- Memcached（テキストプロトコル対応）

## バージョン

0.5
