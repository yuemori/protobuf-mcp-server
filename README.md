# Protobuf MCP Server

Protocol Buffersファイル（.proto）を解析し、セマンティックな情報を提供するMCP（Model Context Protocol）サーバー。

## 目的

- LLMがRAG検索を行わずに正確なProtocol Buffers情報を取得
- トークン消費量を削減してコンテキスト効率を向上
- Claude CodeでのProtobuf開発支援

## 機能

- プロジェクト設定の管理
- Protobuf サービス一覧の取得
- スキーマ詳細情報の提供
- パターン検索機能

## 開発状況

- [x] プロジェクト初期化
- [x] 基本構造作成
- [ ] CLI initコマンド実装
- [ ] activate_projectツール実装
- [ ] テストデータ作成
- [ ] MCP tools実装

## セットアップ

```bash
# 依存関係のインストール
go mod tidy

# プロジェクト初期化
go run cmd/protobuf-mcp/main.go init
```

## アーキテクチャ

Serena MCP Serverの設計思想を参考に、protocompileライブラリを使用してProtocol Buffersファイルの解析を行います。