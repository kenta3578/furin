# furin - Project Instructions

## Overview

ターミナルで動く、名前付き並列インターバルリマインダー TUI ツール。
Go + Bubble Tea で実装。複数タイマーを **1 秒周期の tick 1 本**（`tick()` → `tickMsg`）の上で `Update` がループして進める（goroutine ではない）。YAML 設定で保存・再利用できる。

## Tech Stack

- Go
- Bubble Tea (TUI フレームワーク)
- Lip Gloss (スタイリング)
- go-yaml (設定ファイル読み書き)

## Directory Structure

```
├── ai_docs/       # AI用ドキュメント（北極星・戦略）
├── local_docs/    # ローカル専用（git管理外）
├── main.go        # エントリポイント
└── ...
```

## Development Guidelines

- ブランチ戦略: main（本番）/ develop（開発デフォルト）
- Issue → worktree → PR → develop マージ のフローで開発

## Commands

```bash
go run .              # 開発実行（main.go 単体指定だと config.go 等が読まれずビルドエラー）
go build -o furin .   # ビルド
go test ./...         # テスト
```
