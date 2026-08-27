# IACForge

IACForge は、インフラストラクチャを**モデル**として定義し、様々な表現 (YAML・グラフ・ドキュメント・AI エージェント向けインターフェース) を生成するためのフレームワークです。

「**インフラは知識**」という哲学に基づき、オブジェクトモデルを唯一の信頼できる情報源 (single source of truth) とします。YAML・各種出力はすべてモデルの異なる表現であり、モデルが実装ではなくインフラの概念を表現します。

## 主な特徴

- **オブジェクトモデル** — インフラのすべてのオブジェクトを `Entity`、Entity 間の関係を `Relation` として表現
- **オーナーシップ** — 所有権をツリー構造 (Ownership) で表現
- **複数表現の生成** — markdown / mermaid / svg / json などの Renderer
- **クエリ & 検証** — Entity Kind や Relation Type によるフィルタ、整合性チェック
- **拡張可能** — Entity Kinds、Relation Types、Views、Validation Rules をプラグインとして追加可能。AWS 拡張は 45 種類の Entity Kind、12 種類の Relation Type を提供
- **ベンダーニュートラル** — コアモデルはベンダ固有の概念を持たず、Provider がベンダ情報を担う
- **MCP 対応** — AI エージェントが stdio 経由でモデルを操作できる MCP サーバー (30 ツール)

## Requirements

- Go 1.25 以上（ビルドに必要）
- python3（MCP デモ `demo/mcp_demo.py` の実行に必要）
- サードパーティのランタイム依存なし（純粋な Go により単一バイナリ）

## インストール
```bash
git clone https://github.com/MoyashiWithDevice/IACForge.git
```
リポジトリを取得後、`cmd/iacforge` をビルドします。

```bash
# リポジトリのルートで実行
go build -o iacforge ./cmd/iacforge
```

これでリポジトリ直下に `iacforge` バイナリが生成されます。

Go の `go install` で直接インストールすることもできます。

```bash
go install ./cmd/iacforge
```

動作確認:

```bash
./iacforge version
# => iacforge 0.1.0
```

## Demo の実行

`demo/run-demo.sh` が IACForge の全機能を自動で実行します。初回はビルド込みで実行されます。

```bash
./demo/run-demo.sh
```

ビルド済みのバイナリがある場合は `--skip-build` を付けると省けます。

```bash
./demo/run-demo.sh --skip-build
```

デモで実行される内容:

| セクション | 内容 |
|------------|------|
| 1. validate | `demo/core/model.yaml` の検証（スキーマ違反・ダングリング参照のネガティブケース含む） |
| 2. info | ディレクトリスキャンで連携したグラフのサマリー表示 |
| 3. render | markdown / mermaid / json / svg へのレンダリング |
| 4. query | Entity Kind / Relation Type によるフィルタ (text / json / mermaid) |
| 5. AWS extension | ベンダー固有の Kind / Relation Type の検証・クエリ |
| 6. mcp | stdio 上の MCP サーバーで AI エージェント操作 (30 ツール) |

生成されたアーティファクトは `demo/out/` に出力されます。

## クイックスタート

パス引数を省略すると、カレントディレクトリ直下の `.yaml` / `.yml` を再帰スキャンして単一グラフにマージします。ID は全ファイル間でユニークである必要があります。

```bash
# 検証
./iacforge validate demo/core/model.yaml
./iacforge validate models/

# サマリー表示
./iacforge info demo/core/

# レンダリング
./iacforge render demo/core/model.yaml --format markdown
./iacforge render demo/core/model.yaml --format mermaid
./iacforge render demo/core/model.yaml --format json
./iacforge render demo/core/model.yaml --format svg --output out/graph.svg

# クエリ
./iacforge query demo/core/model.yaml --kind vm --format text
./iacforge query demo/core/model.yaml --type depends_on --format json
./iacforge query demo/core/model.yaml --kind application --format mermaid

# MCP サーバー
./iacforge mcp --port 8080
```

各コマンドのオプションは `iacforge`（引数なし）の usage 出力で確認できます。

## CLI コマンド

| コマンド | 説明 |
|----------|------|
| `validate` | YAML インフラモデルの検証 |
| `info` | インフラモデルのサマリー表示 |
| `render` | ビューをアーティファクトにレンダリング (svg, markdown, mermaid, json) |
| `query` | モデルへのクエリ実行 (formats: text, json, mermaid, markdown) |
| `mcp` | MCP サーバー起動 (stdio / SSE) |
| `version` | バージョン表示 |

## ディレクトリ構成

```
.
├── cmd/iacforge/     # CLI エントリポイント
├── demo/             # デモスクリプトとサンプルモデル
├── spec/             # 仕様書（モデルの絶対的参照先）
├── docs/             # 拡張リファレンスなどのドキュメント
├── src/              # コア実装（core, schema, parser, query, renderer など）
└── testdata/
```

## ドキュメント

- `spec/` — 仕様書（00-philosophy, 01-entity, 02-relation, 03-object-model, ...）
- `docs/extension-reference.md` — 拡張システムリファレンス
- `docs/yaml-reference/` — YAML 構文リファレンス
- `AGENTS.md` — 開発ガイドライン

## テスト

```bash
go test ./...
```
