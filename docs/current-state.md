# 現在の実装状態

> 最終更新: 2026-04-16

## プロジェクト概要
YouTubeをファイルシステムのように探索するインタラクティブCLIツール

## 現在のバージョン
- Version: 0.1.0 (開発中)
- Go 1.25+
- YouTube Data API v3

## 実装済み機能

### コマンド
| コマンド | 説明 | 認証 |
|---|---|---|
| `ls` | チャンネルルート: プレイリスト＋動画一覧、プレイリスト内: 動画一覧 | APIキー |
| `cd [ID]` | チャンネル/プレイリスト間の移動。`..` `~` `/` 対応 | - |
| `pwd` | 現在のパス表示 | - |
| `cp VIDEO_ID PLAYLIST_ID` | プレイリストに動画を追加 | OAuth |
| `open ID` | ブラウザで動画/プレイリストを開く（QRコード付き） | - |
| `cat VIDEO_ID` | 動画の詳細情報を表示（説明文・再生数など） | APIキー |
| `mkdir [-m MODE] TITLE` | 新規プレイリスト作成 | OAuth |
| `chmod MODE PLAYLIST_ID` | プレイリストの公開設定変更 | OAuth |
| `rm VIDEO_ID PLAYLIST_ID` | プレイリストから動画を削除 | OAuth |
| `whoami` | 自分のチャンネル情報表示 | APIキー |
| `mv VIDEO_ID SRC_ID DST_ID` | プレイリスト間で動画を移動 | OAuth |
| `rmdir PLAYLIST_ID` | プレイリストを削除 | OAuth |
| `find` | キーワードで動画検索 | APIキー |
| `grep KEYWORD` | ls結果をキーワードで絞り込み | - |
| `tree` | 階層ツリー表示 | APIキー |
| `exit` / Ctrl+C | シェル終了 | - |

### パス構造
- `//` → 登録チャンネル一覧（OAuth必要）
- `/UCxxxx` → チャンネルのプレイリスト＋動画
- `/UCxxxx/PLxxxx` → プレイリスト内の動画

### タブ補完
- `cd` → プレイリストID一覧
- `cp` → 第1引数: 動画ID、第2引数: プレイリストID
- `rm` → 第1引数: 動画ID、第2引数: プレイリストID
- `mv` → 第1引数: 動画ID、第2-3引数: プレイリストID
- `rmdir` → プレイリストID一覧
- `open` → 動画ID/プレイリストID一覧
- `cat` → 動画ID一覧
- `chmod` → プレイリストID一覧
- `grep` → ls結果のエントリ一覧
- `find` → 検索オプション（-k, -s, -c, -a, -q）

### セットアップ
- 初回起動時の対話セットアップ → `~/.ysh/.env` に保存
- 手動設定: `~/.ysh/.env` に `KEY=VALUE` 形式
- 環境変数 `export` にも対応
- OAuth トークンは `~/.ysh/token.json` にキャッシュ

### 表示
- `d`（青）= チャンネル、`d`（シアン）= プレイリスト、`-`（緑）= 動画
- APIキーのみ → read-only モードで起動、機能制限を明示

## ファイル構成

```
youtube-cli/
├── cmd/ysh/        # ソースコード（package main）
│   ├── main.go     # エントリポイント、App struct、executor/completer
│   ├── cmd_ls.go   # ls系コマンド
│   ├── cmd_cd.go   # cd コマンド
│   ├── cmd_cat.go  # cat コマンド
│   ├── cmd_open.go # open コマンド、QRコード表示
│   ├── cmd_write.go # 書き込み系コマンド
│   ├── cmd_find.go # find コマンド（動画検索）
│   ├── cmd_grep.go # grep コマンド（ls結果絞り込み）
│   ├── cmd_tree.go # tree コマンド（階層表示）
│   ├── cmd_whoami.go # whoami コマンド
│   ├── helpers.go  # ユーティリティ
│   ├── auth.go     # OAuth 2.0 認証フロー
│   └── setup.go    # 初回セットアップ
├── build/          # ビルド成果物（gitignore済み）
├── go.mod
├── build/          # ビルド成果物（gitignore済み）
├── .gitignore
├── AGENTS.md
├── CLAUDE.md
├── README.md
└── docs/
    └── current-state.md
```

## 未実装・今後の候補
- なし（find/grep/tree は実装済み）

## 既知の注意点
- `Subscriptions.List().Mine(true)` は OAuth 必須。APIキーのみでは案内メッセージ表示
- `go-prompt` は PTY 必要なため非対話環境（CI等）では実行不可
- YouTube API の 1日クォータに注意（ls 1回で2-3 API呼び出し）
- OAuth トークンのリフレッシュが自動で行われるが、期限切れ等のエラーハンドリングが未整備