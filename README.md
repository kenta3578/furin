<!-- demo gif here -->

# furin

🔔 ターミナルで動く、名前付き並列インターバルリマインダー

座りすぎ防止・水分補給など、複数のリマインダーを同時に動かせる TUI ツール。絵文字・名前・インターバルを自由にカスタマイズして YAML に保存できます。

## 目的

- 物理タイマーを忘れがちな開発者が、ターミナル上でリマインダーを管理できるようにする
- 複数のリマインダーを並列で動かす（例: 💧30分ごとに水分補給 + 🪑60分ごとに立つ）
- 名前・絵文字・インターバルを自由に設定し、YAML で保存・再利用

## インストール

Go がインストールされていれば、以下のコマンドで導入できます。

```bash
go install github.com/kenta3578/furin@latest
```

`$GOPATH/bin`（デフォルト `~/go/bin`）に `furin` バイナリが配置されます。PATH が通っていれば、そのまま `furin` で起動できます。

ソースからビルドする場合:

```bash
git clone https://github.com/kenta3578/furin.git
cd furin
go build -o furin .
./furin
```

## 使い方

```bash
furin
```

起動するとリマインダー一覧が表示され、各タイマーが並列でカウントダウンします。`a` で新しいタイマーを追加、`?` でキーバインド一覧を表示できます。

## キーバインド

リスト画面で使えるキー一覧です。

| キー | 動作 |
|------|------|
| `↑` / `k` | 選択を上に移動 |
| `↓` / `j` | 選択を下に移動 |
| `Shift+J` | 選択中のタイマーを下へ並び替え |
| `Shift+K` | 選択中のタイマーを上へ並び替え |
| `a` | タイマーを追加 |
| `d` | 選択中のタイマーを削除 |
| `r` | 選択中のタイマーをリセット |
| `p` | 全タイマーを一時停止 / 再開 |
| `+` / `=` | 効果音の音量を上げる |
| `-` / `_` | 効果音の音量を下げる |
| `?` | ヘルプ表示の切り替え |
| `q` / `Ctrl+C` | 終了 |

タイマー追加フォームでは、種別選択ステップで `←` / `→`（`h` / `l` / `Tab`）で `interval` ⇔ `once` を切り替え、`Enter` で次へ、`Esc` でキャンセルできます。

## 設定ファイル

設定は YAML で永続化されます。パスは次の優先順で決まります。

1. `$XDG_CONFIG_HOME/furin/config.yaml`（`XDG_CONFIG_HOME` が設定されている場合）
2. `~/.config/furin/config.yaml`

ファイルが存在しない場合はデフォルト設定で起動し、変更は終了時に自動保存されます。

### YAML サンプル

```yaml
# 効果音の設定（すべて任意。省略時はデフォルト値）
sound_enabled: true   # 効果音の ON/OFF（省略時 true）
sound_name: Glass     # macOS のシステムサウンド名（省略時 Glass）
sound_volume: 1.0     # 音量 0.0〜2.0（省略時 1.0）

timers:
  # interval タイマー: 一定間隔で繰り返し発火する
  - name: 水分補給
    emoji: 💧
    type: interval
    interval: 30        # 分単位

  - name: 座りすぎ防止
    emoji: 🪑
    type: interval
    interval: 60
    action: 立ち上がってストレッチ   # 完了時に表示するアクション（任意）

  # once タイマー: 指定時刻に1回だけ発火する
  - name: 昼休み
    emoji: 🍱
    type: once
    at: "12:00"         # HH:MM 形式
```

- `type` を省略すると `interval` として扱われます（後方互換）。
- `sound_name` は macOS の `/System/Library/Sounds/` にあるサウンド名（例: `Glass`・`Ping`・`Basso`・`Submarine`）を指定します。存在しない名前を指定した場合はデフォルトの `Glass` にフォールバックします。
- `sound_volume` は `0.0`〜`2.0` の範囲にクランプされます。

## 開発

```bash
go run main.go        # 開発実行
go build -o furin .   # ビルド
go test ./...         # テスト
```

## 構成

```
├── ai_docs/       # AI用ドキュメント
├── local_docs/    # ローカル専用（git管理外）
└── ...
```
