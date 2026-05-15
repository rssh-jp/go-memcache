BIN_DIR := bin

## help: このヘルプを表示する
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'

## build: ライブラリをビルドする
build:
	go build ./...

## test: ユニットテストを実行する（Memcached 不要）
test:
	go test ./...

## test-verbose: ユニットテストを詳細ログ付きで実行する
test-verbose:
	go test -v ./...

## test-integration: 統合テストを実行する（要: localhost:11211 で Memcached 起動）
test-integration:
	go test -tags integration ./...

## test-integration-verbose: 統合テストを詳細ログ付きで実行する
test-integration-verbose:
	go test -tags integration -v ./...

## vet: 静的解析を実行する
vet:
	go vet ./...

## tidy: go.mod / go.sum を整理する
tidy:
	go mod tidy

.PHONY: help build test test-verbose test-integration test-integration-verbose vet tidy
