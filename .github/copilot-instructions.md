# go-memcache — Copilot Instructions

## プロジェクト概要

Memcached へのコネクション管理とデータ操作（Get/Set）を提供する Go ライブラリ。

- `memcache` パッケージ: `net.Conn` を薄くラップし、Memcached テキストプロトコルを実装する低レイヤ接続ライブラリ。
- `memcachemanager` パッケージ（ルート）: `memcache.Connection` をコネクションプールとして管理し、複数の名前付き接続先への並列アクセスを提供する高レイヤライブラリ。

## パッケージ構成

```
go-memcache/
├── memcache/
│   ├── memcache.go       # Connection 構造体・Connect/Get/Set/FlushAll
│   └── memcache_test.go
├── memcachemanager.go    # Initialize / Get / Set・コネクションプール管理
├── memcachemanager_test.go
├── go.mod
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
- `writeString()` はバッファサイズ分割して書き込む
- コマンド末尾は必ず `\r\n`（`newline` 定数）を付与すること

### `memcachemanager` パッケージ（memcachemanager.go）

- `connections` はグローバル変数（`map[string]chan *memcache.Connection`）で管理する
- `Initialize()` で各接続先ごとに `ParallelNum` 本のコネクションをチャネルに積む
- `Get()` は `searchConnection()` 経由で最速のコネクションを取得する（接続先名の指定なし）
  - 最初に `shortTimeOut`（1 ナノ秒）で即時探索 → 見つからなければ `timeOut`（5 秒）で再探索
- `Set()` は接続先名を指定してコネクションを取得する
- `defer` でコネクションを必ずチャネルに戻す（プール返却）

### 新しいコマンドを追加する場合

- `memcache/memcache.go` に `(*Connection).Xxx()` メソッドを追加する
- テキストプロトコルのコマンドフォーマットに従う（[Memcached プロトコル仕様](https://github.com/memcached/memcached/blob/master/doc/protocol.txt) 参照）
- `memcache/memcache_test.go` に対応するテストを追加する
- 必要に応じて `memcachemanager.go` にラッパー関数を追加する
- README.md の API リファレンステーブルも更新する

## テスト・検証コマンド

```bash
# テスト実行（要: localhost:11211 で Memcached 起動）
go test ./...

# ビルド確認
go build ./...

# 静的解析
go vet ./...
```

### テスト前提条件

- `localhost:11211` で Memcached が起動していること
- `TestMain` で `flush_all` を実行してから各テストを動かすため、既存データは消去される

## 依存更新

外部依存なし。標準ライブラリのみ使用のため `go mod tidy` のみで十分。

```bash
go mod tidy
```
