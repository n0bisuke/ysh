# ysh - YouTube Shell

YouTube をファイルシステムのように探索できるインタラクティブ CLI ツール。

プレイリストをディレクトリ、動画をファイルに見立てて `cd` / `ls` で自由に移動できます。

## 必要環境

- Go 1.25+

## セットアップ

### 初回起動での対話セットアップ（おすすめ）

`ysh` を初めて起動すると、必要な値がなければセットアップ画面が立ち上がります。

```
$ ./ysh
=== ysh first-time setup ===
Required values are missing. Let's configure them.

[1/2] YouTube Data API v3 key (required)
  Get one at: https://console.cloud.google.com/apis/credentials
  YOUTUBE_API_KEY: AIzaSy...

[2/2] Your YouTube Channel ID (required)
  Format: UCxxxxxxxxxxxxxxxx
  Find at: https://www.youtube.com/account_advanced
  YOUTUBE_CHANNEL_ID: UCxxxxxxxxxxxxxxxx

--- Optional: OAuth credentials (for 'add' command) ---
  Leave blank to skip (read-only mode).
  Create at: https://console.cloud.google.com/apis/credentials
  Redirect URI: http://localhost:8089/callback

  YOUTUBE_CLIENT_ID: xxxx.apps.googleusercontent.com
  YOUTUBE_CLIENT_SECRET: GOCSPX-xxxx

Configuration saved to /home/you/.ysh/.env
```

入力した値は `~/.ysh/.env` に保存され、次回以降は自動で読み込まれます。

### 手動で設定する場合

```bash
mkdir -p ~/.ysh
cat > ~/.ysh/.env << 'EOF'
YOUTUBE_API_KEY=AIzaSy...
YOUTUBE_CHANNEL_ID=UCxxxxxxxxxxxxxxxx
YOUTUBE_CLIENT_ID=xxxx.apps.googleusercontent.com
YOUTUBE_CLIENT_SECRET=GOCSPX-xxxx
EOF
```

`YOUTUBE_CLIENT_ID` / `YOUTUBE_CLIENT_SECRET` は省略可能です。省略すると読み取り専用モードで起動します。

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

## ビルドと起動

```bash
go build -o ysh .
./ysh
```

## 使い方

シェルを起動すると `yt:/ $` プロンプトが表示されます。

OAuth クレデンシャルあり：
```
ysh - YouTube Shell
Mode: full (read + write)
Commands: ls [-a], cd, add, exit
yt:/ $
```

API キーのみ：
```
ysh - YouTube Shell
Mode: read-only (cp command unavailable)
Set YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET in ~/.ysh/.env to enable write operations.
Commands: ls, cd, exit
yt:/ $
```

### コマンド一覧

| コマンド | 説明 |
|---|---|
| `ls` | ルートならプレイリスト＋動画一覧、プレイリスト内なら動画一覧 |
| `cd ID` | プレイリスト（ID）に移動 |
| `cd ..` | ルートに戻る |
| `cd /` | ルートに戻る |
| `cp VIDEO_ID PLAYLIST_ID` | プレイリストに動画を追加（OAuth 認証が必要） |
| `exit` | シェルを終了 |

### 実行例

```
yt:/ $ ls
T    ID                                     TITLE                          INFO
--------------------------------------------------------------------------------
d    PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0     デュアル                        17 items
-    dQw4w9WgXcQ                            Never Gonna Give You Up         video

yt:/ $ cd PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0
yt:/PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0 $ cp dQw4w9WgXcQ PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0
Added video dQw4w9WgXcQ to playlist PLsicJn3Q6ycy9o57w4wQIvTHaLiMmSlO0
```

### cp コマンドと OAuth 認証

`cp` コマンドは初回実行時にブラウザで OAuth 認証画面が開きます。認証後にトークンが `~/.ysh/token.json` に保存され、以降は自動で再利用されます。

## 使用ライブラリ

| ライブラリ | 用途 |
|---|---|
| [google.golang.org/api/youtube/v3](https://pkg.go.dev/google.golang.org/api/youtube/v3) | YouTube Data API v3 クライアント |
| [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2) | OAuth 2.0 認証 |
| [github.com/c-bata/go-prompt](https://github.com/c-bata/go-prompt) | インタラクティブプロンプト |
| [github.com/joho/godotenv](https://github.com/joho/godotenv) | .env ファイルの読み込み |

## ライセンス

MIT