# AGENTS.md

## プロジェクト概要
- ysh: YouTubeをファイルシステムのように探索するインタラクティブCLIツール
- 現在の目標: Unixコマンドライクな操作でYouTubeプレイリスト/動画を閲覧・管理できるシェルの構築
- ユーザーが明示した要件の範囲で変更し、要件を勝手に追加しないこと。

## 作業ルール
- 実装や設計を進める前に、まずユーザーの要件を正確に把握すること。
- 実装に影響する前提が不足している場合は、先にヒアリングすること。
- 推測ベースの大きなリファクタではなく、段階的な実装を優先すること。
- ユーザーから依頼がない限り、既存の方向性を勝手に変えないこと。
- 引き継ぎで開発を再開する際は、前回の課題と今回触る範囲を最初に短く確認してから着手すること。
- 着手後の軽微なバグ調整や切り分けは逐一許可待ちにせず、小さく直して検証まで進めること。
- ユーザー要件を広げる変更は避けること。特に未依頼の機能追加は行わないこと。

## コード方針
- Go 1.25+ を使う。
- 1ファイルあたり約400行以下を目安にすること。大きくなったら責務ごとに分割する。
- 外部パッケージ: google.golang.org/api/youtube/v3, c-bata/go-prompt, joho/godotenv, golang.org/x/oauth2
- コメントは最小限にし、非自明な処理にだけ付けること。
- 秘密情報（APIキー、OAuthトークン等）はソースコードにハードコードせず ~/.ysh/.env に保存すること。
- APIキーのみの read-only モードと OAuth 有りの full モードの2段階に対応すること。
- go-prompt はPTYを要求するため、非対話環境ではテスト不可。ビルド確認のみ行うこと。

## ファイル構成

```
youtube-cli/
├── cmd/ysh/        # ソースコード（package main）
│   ├── main.go     # エントリポイント、App struct、executor/completer
│   ├── cmd_ls.go   # ls系コマンド
│   ├── cmd_cd.go   # cd コマンド
│   ├── cmd_cat.go  # cat コマンド
│   ├── cmd_open.go # open コマンド、QRコード表示
│   ├── cmd_write.go # 書き込み系コマンド（cp, mkdir, chmod, rm, mv, rmdir, ensureOAuth）
│   ├── cmd_whoami.go # whoami コマンド
│   ├── helpers.go  # ユーティリティ（parsePath, filterByPrefix, truncate, formatDuration, formatNumber）
│   ├── auth.go     # OAuth 2.0 認証フロー、トークン保存/読み込み
│   └── setup.go    # 初回セットアップ、設定ファイル読み書き、パス管理
├── go.mod          # モジュール定義 (module ysh)
├── build/          # ビルド成果物（gitignore済み）
├── .env            # 開発用ローカル設定（gitignore済み）
├── .gitignore
├── AGENTS.md
├── CLAUDE.md
├── README.md
└── docs/
    └── current-state.md
```

### 分割ルール
- コマンド実装は `cmd_*.go` ファイルに分割する。
- 書き込み系コマンド（OAuth必要）は `cmd_write.go` にまとめる。
- 読み取り系コマンドは個別ファイルに分割する。
- 新しいコマンド追加時は対応する `cmd_*.go` と main.go の executor/completer を更新する。

## パス構造
- `//` → 登録チャンネル一覧（OAuth必要、APIキーのみでは案内メッセージ表示）
- `/UCxxxx` → チャンネルのプレイリスト＋動画一覧
- `/UCxxxx/PLxxxx` → プレイリスト内の動画一覧
- 特殊: `cd ~` で自分のチャンネルに戻る、`cd /` でチャンネル一覧に戻る

## ログ・記録ルール
「ログに残して」「記録して」などの指示があった場合、以下を更新すること：
1. `docs/current-state.md` — 実装状態サマリ（必須）
2. `AGENTS.md` のステータス欄 — ステータスが大きく変わった場合のみ

## 現在のステータス
- フェーズ: 基本機能実装完了、追加コマンド拡張中
- 詳細な実装状態は docs/current-state.md を参照

## 検証
- コード変更後は `go build -o build/ysh ./cmd/ysh/` でビルドを確認すること。
- go-prompt は非対話環境で動作しないため、実行時テストはユーザー環境で行うこと。
- 未実装の部分がある場合は、その点を明確に伝えること。