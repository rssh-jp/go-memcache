BIN_DIR := bin

## help: このヘルプを表示する
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'

## build: ライブラリをビルドする
build:
	go build ./...

## test: ユニットテストを実行する（要: localhost:11211 で Memcached 起動）
test:
	go test ./...

## test-verbose: ユニットテストを詳細ログ付きで実行する
test-verbose:
	go test -v ./...

## vet: 静的解析を実行する
vet:
	go vet ./...

## tidy: go.mod / go.sum を整理する
tidy:
	go mod tidy

.PHONY: help build test test-verbose vet tidy
