# YAML Reference

IACForgeのYAMLファイル作成のための完全版リファレンスです。

## 目次

| ドキュメント | 内容 |
|--------------|------|
| [Document Structure](document-structure.md) | YAMLドキュメントの基本構造 |
| [Entity Syntax](entity-syntax.md) | Entity共通プロパティ、ステータス値 |
| [Entity Kinds](entity-kinds.md) | 全Entity種類の定義とプロパティ |
| [Relation Syntax](relation-syntax.md) | Relation共通構文、participantフォーマット |
| [Relation Types](relation-types.md) | 全Relation種類の定義とプロパティ |
| [References](references.md) | 参照構文（シンプル、修飾、インターフェース） |
| [Validation](validation.md) | 検証ルール、命名規則、Graph制約 |
| [Example](example.md) | 完全なインフラモデルの例 |

## 共通ルール

### 必須フィールド

| Object | Required Fields |
|--------|-----------------|
| Entity | id, kind, name |
| Nested Entity | id (自動生成可能) |
| Relation | id, type, participants |

### ネストルール

| 親 Kind | ネストキー | 子 Kind |
|---------|-----------|---------|
| site | racks | rack |
| rack | servers | server |
| server | networks | network |
| server | vms | vm |
| network | interfaces | interface |
| vm | applications | application |

### ステータス値

| Value | Description |
|-------|-------------|
| planned | Not yet deployed |
| active | Operational |
| maintenance | Under maintenance |
| deprecated | Scheduled for removal |
| offline | Not operational |

### 命名規則

| 対象 | パターン | 例 |
|------|----------|-----|
| ID | kebab-case (推奨) | `srv-proxmox-01` |
| Kind | 小文字・単数形 | `server`, `vm` |
| Relation Type | snake_case | `connects`, `depends_on` |
| Property | snake_case | `cpu_cores`, `memory_gb` |
