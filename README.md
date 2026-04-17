# ysh - YouTube Shell

YouTube をファイルシステムのように探索できるインタラクティブ CLI ツール。

プレイリストをディレクトリ、動画をファイルに見立てて `cd` / `ls` で自由に移動できます。

## 必要環境

- Go 1.25+（ソースからビルドする場合）

## インストール

### バイナリダウンロード（おすすめ）

OS とアーキテクチャを自動判定して最新版をインストール：

```bash
curl -fsSL https://raw.githubusercontent.com/n0bisuke/youtube-cli/main/install.sh | sh
```

### GitHub Releases から手動ダウンロード

[Releases ページ](https://github.com/n0bisuke/youtube-cli/releases) から該当バイナリをダウンロード：

| ファイル | OS | アーキテクチャ |
|---|---|---|
| `ysh-darwin-arm64` | macOS | Apple Silicon (M1/M2/M3) |
| `ysh-darwin-amd64` | macOS | Intel |
| `ysh-linux-arm64` | Linux | ARM64 |
| `ysh-linux-amd64` | Linux | x86_64 |
| `ysh-windows-amd64.exe` | Windows | x86_64 |

```bash
# 例: macOS Apple Silicon
curl -fsSL -o /usr/local/bin/ysh https://github.com/n0bisuke/youtube-cli/releases/latest/download/ysh-darwin-arm64
chmod +x /usr/local/bin/ysh
```

### ソースからビルド

```bash
git clone https://github.com/n0bisuke/youtube-cli.git
cd youtube-cli
go build -ldflags "-X main.version=$(git describe --tags --always)" -o build/ysh ./cmd/ysh/
```

## セットアップ

### 初回起動での対話セットアップ（おすすめ）

`ysh` を初めて起動すると、必要な値がなければセットアップ画面が立ち上がります。

**API キー + OAuth の場合:**
```
$ ./ysh
=== ysh first-time setup ===

You need either an API key (read-only) or OAuth credentials (full access).

[1/2] Your YouTube Channel ID (required)
  Format: UCxxxxxxxxxxxxxxxx
  Find at: https://www.youtube.com/account_advanced
  YOUTUBE_CHANNEL_ID: UCxxxxxxxxxxxxxxxx

[2/2] YouTube Data API v3 key (optional — skip if using OAuth)
  Get one at: https://console.cloud.google.com/apis/credentials
  YOUTUBE_API_KEY (press Enter to skip): AIzaSy...

--- Optional: OAuth credentials (for write operations) ---
  Leave blank to skip (read-only mode with API key).
  Create at: https://console.cloud.google.com/apis/credentials
  Redirect URI: http://localhost:8089/callback

  YOUTUBE_CLIENT_ID: xxxx.apps.googleusercontent.com
  YOUTUBE_CLIENT_SECRET: GOCSPX-xxxx

Configuration saved to /home/you/.ysh/.env
```

**OAuth のみの場合（APIキー不要）:**
```
[2/2] YouTube Data API v3 key (optional — skip if using OAuth)
  YOUTUBE_API_KEY (press Enter to skip):

--- OAuth credentials (required when API key is not set) ---
  YOUTUBE_CLIENT_ID: xxxx.apps.googleusercontent.com
  YOUTUBE_CLIENT_SECRET: GOCSPX-xxxx
```

入力した値は `~/.ysh/.env` に保存され、次回以降は自動で読み込まれます。

### 手動で設定する場合

**APIキーのみ（読み取り専用）:**
```bash
mkdir -p ~/.ysh
cat > ~/.ysh/.env << 'EOF'
YOUTUBE_API_KEY=AIzaSy...
YOUTUBE_CHANNEL_ID=UCxxxxxxxxxxxxxxxx
EOF
```

**OAuthのみ（フルアクセス）:**
```bash
mkdir -p ~/.ysh
cat > ~/.ysh/.env << 'EOF'
YOUTUBE_CHANNEL_ID=UCxxxxxxxxxxxxxxxx
YOUTUBE_CLIENT_ID=xxxx.apps.googleusercontent.com
YOUTUBE_CLIENT_SECRET=GOCSPX-xxxx
EOF
```

**両方（推奨）:**
```bash
mkdir -p ~/.ysh
cat > ~/.ysh/.env << 'EOF'
YOUTUBE_API_KEY=AIzaSy...
YOUTUBE_CHANNEL_ID=UCxxxxxxxxxxxxxxxx
YOUTUBE_CLIENT_ID=xxxx.apps.googleusercontent.com
YOUTUBE_CLIENT_SECRET=GOCSPX-xxxx
EOF
```

### 環境変数で渡す場合

シェルで `export` しても認識されます。

```bash
export YOUTUBE_API_KEY=AIzaSy...
export YOUTUBE_CHANNEL_ID=UCxxxxxxxxxxxxxxxx
./ysh
```

### 設定の読み込み優先順位

1. 環境変数（`export` 済みのもの）
2. `~/.ysh/.env`（グローバル設定）
3. `./.env`（カレントディレクトリ、開発用）

## 各値の取得方法

### YouTube API キー

1. [Google Cloud Console](https://console.cloud.google.com/) にログイン
2. プロジェクトを作成（既存のものがあれば選択）
3. 左メニューの **API とサービス → ライブラリ** を開く
4. **YouTube Data API v3** を検索して **有効化**
5. 左メニューの **API とサービス → 認証情報** を開く
6. 上部の **認証情報を作成 → API キー** をクリック
7. 表示されたキーをコピーする
   - 必要に応じてキー制限で「YouTube Data API v3」のみ許可しておくと安全

### チャンネル ID

1. YouTube で自分のチャンネルを開く
2. URL が `https://www.youtube.com/channel/UCxxxxxxxxxxxxxxxx` なら `UC...` の部分がチャンネル ID
   - `https://www.youtube.com/@handle` 形式の場合は、[このページ](https://www.youtube.com/account_advanced)の **チャンネル ID** から確認可能

### OAuth 2.0 クライアント ID（cp コマンドを使う場合のみ）

1. Google Cloud Console の **API とサービス → 認証情報** を開く
2. **認証情報を作成 → OAuth クライアント ID** をクリック
3. アプリケーションの種類は **デスクトップ アプリ** を選択
4. 名前を入力して作成
5. クライアント ID とクライアント シークレットをメモ
6. 作成した OAuth クライアントの **承認済みリダイレクト URI** に `http://localhost:8089/callback` を追加

## ビルドと起動（開発用）

```bash
go build -o build/ysh ./cmd/ysh/
./build/ysh
```

## 使い方

シェルを起動するとプロンプトが表示されます。

OAuth クレデンシャルあり：
```
ysh 0.1.0 - YouTube Shell
Loading...
...
Mode: full (API key + OAuth)
Commands: ls, cd, cat, cp, mkdir, chmod, rm, mv, rmdir, whoami, open, pwd, exit
yt:/UCxxxx $
```

OAuth のみ（APIキーなし）：
```
ysh 0.1.0 - YouTube Shell
Loading...
...
Mode: full (OAuth)
Commands: ls, cd, cat, cp, mkdir, chmod, rm, mv, rmdir, whoami, open, pwd, exit
yt:/UCxxxx $
```

API キーのみ：
```
ysh 0.1.0 - YouTube Shell
Loading...
...
Mode: read-only (API key)
Set YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET in ~/.ysh/.env to enable write operations.
Commands: ls, cd, cat, whoami, open, pwd, exit
yt:/UCxxxx $
```

### パス構造

| パス | 内容 | 認証 |
|---|---|---|
| `//` | 登録チャンネル一覧 | OAuth |
| `/UCxxxx` | チャンネルのプレイリスト＋動画 | APIキー |
| `/UCxxxx/PLxxxx` | プレイリスト内の動画 | APIキー |

### コマンド一覧

| コマンド | 説明 | 認証 |
|---|---|---|
| `ls` | 現在の位置の内容を一覧表示 | APIキー |
| `cd ID` | チャンネル/プレイリストに移動 | - |
| `cd ..` | 上の階層に戻る | - |
| `cd ~` | 自分のチャンネルに戻る | - |
| `pwd` | 現在のパスを表示 | - |
| `cat VIDEO_ID` | 動画の詳細情報を表示 | APIキー |
| `cp VIDEO_ID PLAYLIST_ID` | プレイリストに動画を追加 | OAuth |
| `mkdir [-m MODE] TITLE` | 新規プレイリスト作成（デフォルト: unlisted） | OAuth |
| `chmod MODE PLAYLIST_ID` | プレイリストの公開設定変更（public/unlisted/private） | OAuth |
| `rm VIDEO_ID PLAYLIST_ID` | プレイリストから動画を削除 | OAuth |
| `mv VIDEO_ID SRC_ID DST_ID` | プレイリスト間で動画を移動 | OAuth |
| `rmdir PLAYLIST_ID` | プレイリストを削除 | OAuth |
| `whoami` | 自分のチャンネル情報を表示 | APIキー |
| `open ID` | ブラウザで動画/プレイリストを開く（QRコード付き） | - |
| `exit` / Ctrl+C | シェルを終了 | - |

### タブ補完

- 入力済みのプレフィックスでフィルタリング（例: `open 9` → `9`から始まるIDのみ）
- 候補に `[PL]` / `[V]` プレフィックスでプレイリスト/動画を判別
- `cp` / `rm` / `mv` は引数位置に応じて切り替え（動画ID/プレイリストID）
- `rmdir` / `chmod` はプレイリストIDを補完

### 実行例

```
yt:/UCxxxx $ ls
T    ID                                     TITLE                          INFO
--------------------------------------------------------------------------------
d    PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0     デュアル                        17 items
-    9tv05ehXkrM                            動画タイトル                     video

yt:/UCxxxx $ cd PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0
yt:/UCxxxx/PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0 $ ls
VIDEO_ID       TITLE                                         POSITION
--------------------------------------------------------------------------------
9tv05ehXkrM    動画タイトル                                     0

yt:/UCxxxx/PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0 $ open 9tv05ehXkrM
https://www.youtube.com/watch?v=9tv05ehXkrM
████▀▀██▀▀████ ...
Opened in browser.

yt:/UCxxxx/PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0 $ cp 9tv05ehXkrM PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0
Added video 9tv05ehXkrM to playlist PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0

yt:/UCxxxx/PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0 $ cd ..
yt:/UCxxxx $ cd ..
yt:// $
```

### cp コマンドと OAuth 認証

`cp` コマンドは初回実行時にブラウザで OAuth 認証画面が開きます。認証後にトークンが `~/.ysh/token.json` に保存され、以降は自動で再利用されます。

## 対応予定のコマンド（TODO）

| コマンド | YouTube 操作 | 例 | 優先度 |
|---|---|---|---|
| `find` | 動画を検索 | `find キーワード` | 高 |
| `grep` | 現在の ls 結果からタイトルで絞り込み | `grep キーワード` | 低 |
| `tree` | チャンネルのプレイリスト/動画をツリー表示 | `tree` | 低 |

## 使用ライブラリ

| ライブラリ | 用途 |
|---|---|
| [google.golang.org/api/youtube/v3](https://pkg.go.dev/google.golang.org/api/youtube/v3) | YouTube Data API v3 クライアント |
| [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2) | OAuth 2.0 認証 |
| [github.com/c-bata/go-prompt](https://github.com/c-bata/go-prompt) | インタラクティブプロンプト |
| [github.com/joho/godotenv](https://github.com/joho/godotenv) | .env ファイルの読み込み |
| [github.com/mdp/qrterminal](https://github.com/mdp/qrterminal) | QRコード生成 |

## ライセンス

MIT