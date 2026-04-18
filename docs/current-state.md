# 現在の実装状態

> 最終更新: 2026-04-17

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
| `cp VIDEO_ID PLAYLIST_ID [POSITION]` | プレイリストに動画を追加（位置指定可） | OAuth |
| `open ID` | ブラウザで動画/プレイリストを開く（QRコード付き） | - |
| `cat VIDEO_ID` | 動画の詳細情報を表示（説明文・再生数など） | APIキー |
| `mkdir [-m MODE] TITLE` | 新規プレイリスト作成 | OAuth |
| `chmod MODE PLAYLIST_ID` | プレイリストの公開設定変更 | OAuth |
| `rm VIDEO_ID PLAYLIST_ID` | プレイリストから動画を削除 | OAuth |
| `whoami` | 自分のチャンネル情報表示 | APIキー |
| `mv VIDEO_ID SRC_ID DST_ID [POSITION]` | プレイリスト間で動画を移動（位置指定可） | OAuth |
| `rmdir PLAYLIST_ID` | プレイリストを削除 | OAuth |
| `find` | キーワードで動画検索（`--live`/`--event-type`で配信フィルター可） | APIキー（`-a`でOAuth） |
| `grep KEYWORD` | ls結果をキーワードで絞り込み | - |
| `tree` | 階層ツリー表示 | APIキー |
| `sort --by title\|date PLAYLIST_ID` | プレイリストの動画を並び替え（`-r`で逆順） | OAuth |
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
- `find` → 検索オプション（-k, -s, -c, -a, -e, -l, -q）
- `sort` → オプション（-b, -r）

### セットアップ
- 初回起動時の対話セットアップ → `~/.ysh/.env` に保存
- 手動設定: `~/.ysh/.env` に `KEY=VALUE` 形式
- 環境変数 `export` にも対応
- OAuth トークンは `~/.ysh/token.json` にキャッシュ
- CI環境では `YOUTUBE_TOKEN_JSON` 環境変数でトークン渡し可能

### 表示
- `d`（青）= チャンネル、`d`（シアン）= プレイリスト、`-`（緑）= 動画
- APIキーのみ → read-only モードで起動、機能制限を明示

## ファイル構成

```
youtube-cli/
├── cmd/ysh/           # ソースコード（package main）
│   ├── main.go        # エントリポイント、App struct、executor/completer
│   ├── cmd_ls.go      # ls系コマンド
│   ├── cmd_cd.go      # cd コマンド
│   ├── cmd_cat.go     # cat コマンド
│   ├── cmd_open.go    # open コマンド、QRコード表示
│   ├── cmd_write.go   # 書き込み系コマンド（cp, mkdir, chmod, rm, mv, rmdir）
│   ├── cmd_find.go    # find コマンド（動画検索、配信フィルター）
│   ├── cmd_grep.go    # grep コマンド（ls結果絞り込み）
│   ├── cmd_tree.go    # tree コマンド（階層表示）
│   ├── cmd_sort.go    # sort コマンド（プレイリスト並び替え）
│   ├── cmd_whoami.go  # whoami コマンド
│   ├── helpers.go     # ユーティリティ
│   ├── auth.go        # OAuth 2.0 認証フロー（YOUTUBE_TOKEN_JSON対応）
│   └── setup.go       # 初回セットアップ
├── build/             # ビルド成果物（gitignore済み）
├── go.mod
├── .github/workflows/
│   ├── release.yml        # リリース用（タグpush時）
│   └── find-and-add.yml   # 定期検索＆プレイリスト追加（30分ごと）
├── install.sh             # インストールスクリプト
├── AGENTS.md
├── CLAUDE.md
├── README.md
└── docs/
    └── current-state.md
```

## 未実装・今後の候補
- なし（find/grep/tree/sort は実装済み）

## 既知の注意点
- `Subscriptions.List().Mine(true)` は OAuth 必須。APIキーのみでは案内メッセージ表示
- `go-prompt` は PTY 必要なため非対話環境（CI等）では実行不可。ただし `ysh <command>` のワンショットモードはCIで動作する
- YouTube API の 1日クォータに注意（ls 1回で2-3 API呼び出し）
- OAuth トークンのリフレッシュが自動で行われるが、期限切れ等のエラーハンドリングが未整備
- `sort` コマンドは各アイテムのpositionを1つずつ更新するため、50アイテムで最大50回のAPI呼び出しが発生

## CI/自動化
- ワンショットモード: `ysh find -k KEYWORD -s 1h -a PLxxxx -q`
- CI環境では `YOUTUBE_TOKEN_JSON` 環境変数でOAuthトークンを渡す（`~/.ysh/token.json` の内容をJSON文字列として設定）
- GitHub Actions ワークフロー: `.github/workflows/find-and-add.yml`（30分ごとに定期実行）
- 必要なSecrets: `YOUTUBE_API_KEY`, `YOUTUBE_CHANNEL_ID`, `YOUTUBE_CLIENT_ID`, `YOUTUBE_CLIENT_SECRET`, `YOUTUBE_TOKEN_JSON`

## 今回のセッションでの修正・追加（2026-04-17）
- main.go の completer 関数修正（`case "find":` が大量重複していた破損を修復）
- cmd_open.go: `cmd.Start()` → `cmd.Run()`（ゾンビプロセス防止）
- cmd_open.go: `isPlaylistID()` に `PL` プレフィクスフォールバック追加
- find コマンド: `--live`/`--event-type` オプション追加（配信フィルター）
- grep コマンド: 新規実装（ls結果のローカルフィルタリング）
- tree コマンド: 新規実装（階層ツリー表示、API呼び出しあり）
- sort コマンド: 新規実装（プレイリスト並び替え、title/date、-rで逆順）
- cp/mv コマンド: 位置指定引数（第3/第4引数）対応
- auth.go: `YOUTUBE_TOKEN_JSON` 環境変数対応（CI用トークン復元）
- GitHub Actions: `.github/workflows/find-and-add.yml` 追加