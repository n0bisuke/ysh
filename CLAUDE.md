# CLAUDE.md
@AGENTS.md
@docs/current-state.md

## Claude Code 固有ルール
- 日本語でコミュニケーションすること
- ファイル読み込みは必要な範囲だけにすること（offset/limitを活用してトークンを節約）
- 大きなコード探索はAgent（サブエージェント）に任せ、メインコンテキストを節約すること
- 実装前に EnterPlanMode で方針を合わせること（1行修正レベルは除く）
- sudo は使用しないこと

## プロジェクト固有の注意点
- go-prompt は PTY を要求するため、`go build` でのビルド確認のみ可能（実行テストは不可）
- Edit ツールでタブインデントファイルを編集する際、スペースインデントの指定だとマッチしないことがある。必要に応じて `sed -i` で挿入すること
- `google.golang.org/api` の API は `option.WithAPIKey()` と `option.WithHTTPClient()` の2つの認証方式を使い分ける
- `Mine(true)` を使う API 呼び出し（Subscriptions等）は OAuth 認証が必須。APIキーのみでは401エラーになるため、事前にチェックして案内メッセージを表示すること