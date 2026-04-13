# ysh - YouTube Shell

YouTube をファイルシステムのように探索できるインタラクティブ CLI ツール。

プレイリストをディレクトリ、動画をファイルに見立てて `cd` / `ls` で自由に移動できます。

## 必要環境

- Go 1.25+
- YouTube Data API v3 の API キー

## セットアップ

### 1. YouTube API キーの取得

1. [Google Cloud Console](https://console.cloud.google.com/) にログイン
2. プロジェクトを作成（既存のものがあれば選択）
3. 左メニューの **API とサービス → ライブラリ** を開く
4. **YouTube Data API v3** を検索して **有効化**
5. 左メニューの **API とサービス → 認証情報** を開く
6. 上部の **認証情報を作成 → API キー** をクリック
7. 表示されたキーをコピーする
   - 必要に応じてキー制限で「YouTube Data API v3」のみ許可しておくと安全

### 2. チャンネル ID の確認

1. YouTube で自分のチャンネルを開く
2. URL が `https://www.youtube.com/channel/UCxxxxxxxxxxxxxxxx` なら `UC...` の部分がチャンネル ID
   - `https://www.youtube.com/@handle` 形式の場合は、[このページ](https://www.youtube.com/account_advanced)の **チャンネル ID** から確認可能

### 3. .env ファイルの作成

プロジェクトルートに `.env` を作成し、API キーとチャンネル ID を記述します。

```
YOUTUBE_API_KEY=AIzaSy...（あなたのAPIキー）
YOUTUBE_CHANNEL_ID=UCxxxxxxxxxxxxxxxx
```

### 4. ビルドと起動

```bash
go build -o ysh .
./ysh
```

## 使い方

シェルを起動すると `yt:/ $` プロンプトが表示されます。

```
ysh - YouTube Shell
Type ls, cd, or exit.
yt:/ $
```

### コマンド一覧

| コマンド | 説明 |
|---|---|
| `ls` | 現在の位置の内容を一覧表示 |
| `cd ID` | プレイリスト（ID）に移動 |
| `cd ..` | ルートに戻る |
| `cd /` | ルートに戻る |
| `exit` | シェルを終了 |

### 実行例

```
yt:/ $ ls
ID                                       TITLE                           ITEMS
--------------------------------------------------------------------------------
PLx0sYbCqOb8TBPRdmBH...  Go Tutorial                     12
PLx0sYbCqOb8SK5...       Web Dev Playlist                8

yt:/ $ cd PLx0sYbCqOb8TBPRdmBH...
yt:/PLx0sYbCqOb8TBPRdmBH... $ ls
VIDEO_ID       TITLE                                         POSITION
--------------------------------------------------------------------------------
dQw4w9WgXcQ    Never Gonna Give You Up                       0
abcdef12345    Go Basics: Variables                          1

yt:/PLx0sYbCqOb8TBPRdmBH... $ cd ..
yt:/ $ exit
Bye!
```

## 使用ライブラリ

| ライブラリ | 用途 |
|---|---|
| [google.golang.org/api/youtube/v3](https://pkg.go.dev/google.golang.org/api/youtube/v3) | YouTube Data API v3 クライアント |
| [github.com/c-bata/go-prompt](https://github.com/c-bata/go-prompt) | インタラクティブプロンプト |
| [github.com/joho/godotenv](https://github.com/joho/godotenv) | .env ファイルの読み込み |

## ライセンス

MIT