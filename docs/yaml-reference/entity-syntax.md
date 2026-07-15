# Entity Syntax

[← README](README.md)

---

## 必須プロパティ

| Property | Type | Description |
|----------|------|-------------|
| id | string | ユニーク識別子（kebab-case推奨） |
| kind | string | Entity kind（小文字・単数形） |
| name | string | 人間が読める名前 |

## 任意の共通プロパティ（`attributes:` 配下）

必須プロパティ以外の共通プロパティは、`attributes:` サブキーの配下に配置します。

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| owner | string | - | 親Entity ID（所有権階層） |
| description | string | - | ドキュメント（Markdown対応） |
| status | enum | - | ライフサイクル状態 |
| tags | list[string] | - | グループ用ラベル |
| labels | map[string] | - | マシン可読メタデータ |
| extensions | map[string] | - | 拡張データ |

## ステータス値

| Value | Description |
|-------|-------------|
| planned | 未デプロイ |
| active | 稼働中 |
| maintenance | メンテナンス中 |
| deprecated | 削除予定 |
| offline | オフライン |

## 基本的なEntity

```yaml
- id: site-tokyo-01
  kind: site
  name: Tokyo Datacenter 1
```

## 全プロパティを指定したEntity

```yaml
- id: srv-proxmox-01
  kind: server
  name: Proxmox Node 01
  attributes:
    description: "Primary Proxmox server"
    status: active
    tags:
      - production
      - compute
    labels:
      region: ap-northeast-1
      environment: production
    extensions:
      vendor: dell
      model: r740xd
  spec:
    platform: proxmox
    cpu_cores: 32
    memory_gb: 128
    storage_gb: 2000
    ip_address: 10.0.1.10
```
