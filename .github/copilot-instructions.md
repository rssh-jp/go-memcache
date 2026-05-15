# go-memcache — Copilot Instructions

## プロジェクト概要

Memcached へのコネクション管理とデータ操作を提供する Go ライブラリ。

- `memcache` パッケージ: `net.Conn` を薄くラップし、Memcached テキストプロトコルの主要コマンドを実装する低レイヤ接続ライブラリ。
- `memcachemanager` パッケージ（ルート）: `memcache.Connection` をコネクションプールとして管理し、複数の名前付き接続先への並列アクセスを提供する高レイヤライブラリ。

## パッケージ構成

```
go-memcache/
├── memcache/
│   ├── memcache.go       # Connection 構造体・全コマンド実装
│   └── memcache_test.go
├── memcachemanager.go    # Initialize / Get / Set / SetEx / Delete・コネクションプール管理
├── memcachemanager_test.go
├── go.mod
├── Makefile
└── README.md
```

## コーディングルール

- Go 1.26 以上の構文・機能を使用する
- 外部依存ライブラリは持たない（標準ライブラリのみ）
- エラーは呼び出し元に返す。`memcache` パッケージ内でのログ出力は禁止
- `memcachemanager` パッケージではタイムアウト時のログ出力は許容する

## 設計上の注意点

### `memcache` パッケージ（memcache/memcache.go）

- Memcached テキストプロトコルで通信する（バイナリプロトコル非対応）
- `bufio.ReadWriter` でバッファリングして送受信する
- `readLine()` は `Flush()` を呼んだ後に ReadLine ループを回す（isPrefix 対応）
- `writeString()` はバッファサイズ分割して書き込みエラーを返す（エラーを無視しない）
- コマンド末尾は必ず `\r\n`（`newline` 定数）を付与すること

#### 実装済みコマンド

| メソッド | Memcached コマンド | 説明 |
|---|---|---|
| `Get(key)` | `get` | 値を取得する |
| `Set(key, value)` | `set` | 値を保存する（exptime=0） |
| `SetEx(key, value, exptime)` | `set` | 有効期限付きで値を保存する |
| `Add(key, value)` | `add` | キーが存在しない場合のみ保存する |
| `Replace(key, value)` | `replace` | キーが存在する場合のみ更新する |
| `Delete(key)` | `delete` | キーを削除する |
| `Increment(key, delta)` | `incr` | 数値を加算する |
| `Decrement(key, delta)` | `decr` | 数値を減算する |
| `FlushAll()` | `flush_all` | 全キーを削除する |

- `Set` / `Add` / `Replace` は内部で `storeCommand()` ヘルパーを共用している
- `Increment` / `Decrement` は内部で `incrDecrCommand()` ヘルパーを共用している
- ストレージ系は `STORED` → `(true, nil)`、`NOT_STORED` → `(false, nil)`、`ERROR` → `(false, err)`
- `Delete` は `DELETED` → `(true, nil)`、`NOT_FOUND` → `(false, nil)`、`ERROR` → `(false, err)`

### `memcachemanager` パッケージ（memcachemanager.go）

- `connections` はグローバル変数（`map[string]chan *memcache.Connection`）で管理する
- `Initialize()` で各接続先ごとに `ParallelNum` 本のコネクションをチャネルに積む
- `Get()` は `searchConnection()` 経由で最速のコネクションを取得する（接続先名の指定なし）
  - 最初に `shortTimeOut`（1 ナノ秒）で即時探索 → 見つからなければ `timeOut`（5 秒）で再探索
- `Set()` / `SetEx()` / `Delete()` は全接続先に操作を適用する（`Get` との整合性を保つため）
- コネクションは取得後に必ずチャネルに戻す（プール返却）。`defer` ではなく処理直後に返却する

### 新しいコマンドを追加する場合

- `memcache/memcache.go` に `(*Connection).Xxx()` メソッドを追加する
- ストレージコマンド（set/add/replace 系）は `storeCommand()` ヘルパーを使い回すこと
- incr/decr 系は `incrDecrCommand()` ヘルパーを使い回すこと
- テキストプロトコルのコマンドフォーマットに従う（[Memcached プロトコル仕様](https://github.com/memcached/memcached/blob/master/doc/protocol.txt) 参照）
- `memcache/memcache_test.go` に対応するテストを追加する
- 必要に応じて `memcachemanager.go` にラッパー関数を追加する（全接続先への適用を忘れずに）
- README.md の API リファレンステーブルも更新する

## テスト・検証コマンド

```bash
# テスト実行（要: localhost:11211 で Memcached 起動）
make test

# 詳細ログ付きテスト
make test-verbose

# ビルド確認
make build

# 静的解析
make vet

# go.mod 整理
make tidy
```

### Makefile ターゲット一覧

| ターゲット | 説明 |
|---|---|
| `make build` | ライブラリをビルドする |
| `make test` | ユニットテストを実行する |
| `make test-verbose` | 詳細ログ付きでテストを実行する |
| `make vet` | 静的解析を実行する |
| `make tidy` | `go.mod` / `go.sum` を整理する |

### テスト前提条件

- `localhost:11211` で Memcached が起動していること
- `TestMain` で `flush_all` を実行してから各テストを動かすため、既存データは消去される

## 依存更新

外部依存なし。標準ライブラリのみ使用のため `go mod tidy` のみで十分。

```bash
make tidy
```
