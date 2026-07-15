# TASK2: ネスト定義（Nested Entity Definition）

## 概要

エンティティの子要素（interface, network等）を、親要素のspec内でネスト定義できるようにする。

## 設計方針

### 1. YAML構文

子要素は親エンティティの`spec:`セクション内で定義する。

```yaml
objects:
  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    spec:
      cpu_cores: 32
      memory_gb: 128
      networks:
        - id: net-private
          name: private
          spec:
            cidr: 172.31.0.0/24
            gateway: 172.31.0.254
          interfaces:
            - id: eth1
              name: proxmox/eth1
              spec:
                ip_address: 172.31.0.15
                type: ethernet

        - name: mgmt
          spec:
            cidr: 10.0.0.0/24
          interfaces:
            - name: eth0
              spec:
                ip_address: 10.0.0.10
```

### 2. ネスト可能ないたわれルール

現在のowner実装（`spec/concrete/14-entity-kinds.md`の"Typical Ownership"）に準拠する。

| 親エンティティ | ネストキー | 子エンティティ |
|---------------|-----------|---------------|
| site | racks | rack |
| site | clusters | cluster |
| rack | servers | server |
| rack | switches | switch |
| rack | routers | router |
| rack | firewalls | firewall |
| server | networks | network |
| server | vms | vm |
| switch | interfaces | interface |
| router | interfaces | interface |
| firewall | interfaces | interface |
| firewall | acls | acl |
| vm | networks | network |
| vm | applications | application |
| network | interfaces | interface |
| application | open_ports | open_port |
| acl | acl_rules | acl_rule |

> **注意**: serverとvmはinterfaceを直接子要素としない。interfaceはnetworkを経由してのみ子要素となる。
> 階層: server > network > interface, vm > network > interface

### 3. 必須項目

ネスト定義内では以下の項目は任意：

| 項目 | 必須 | 備考 |
|------|------|------|
| id | 任意 | 外部参照時にのみ必須 |
| kind | 任意 | ネストキーから自動推定 |
| name | 任意 | 未指定時はIDを使用 |
| spec | 任意 | kind固有のプロパティ |

### 4. ID制約

- IDに`/`（スラッシュ）は使用不可
- スラッシュはパス区切り文字として予約
- 推奨形式：ケバブケース（例: `net-private`, `eth1`）

### 5. 参照方法

#### ID参照（基本）
```yaml
participants:
  source: eth1
  target: sw-core-01/port1
```

#### パス参照（オプション）
```yaml
participants:
  source: srv-proxmox-01/net-private/eth1
  target: sw-core-01/port1
```

#### 参照解決ロジック
1. まず指定されたIDでエンティティを検索
2. IDで見つからない場合、パス表記として解釈
3. パスの末尾セグメントをIDとして検索
4. 親子関係を確認して解決

### 6. 所有権

ネスト構造から自動設定：
- ネストされたエンティティの`owner`は親エンティティのIDに設定
- `owner`フィールドはネスト定義内で指定不要

### 7. 混在

フラット定義（現在の形式）とネスト定義を同一ファイルで混在可能。

```yaml
objects:
  # フラット定義
  - id: rack-a01
    kind: rack
    name: Rack A01
    attributes:
      owner: site-tokyo-01

  # ネスト定義
  - id: srv-proxmox-01
    kind: server
    spec:
      networks:
        - id: net-private
          interfaces:
            - id: eth1
```

## 実装計画

### Phase 1: スキーマ定義
- [ ] `NestingDefinition`構造体の追加
- [ ] `EntityKindDefinition`に`NestingDefs`フィールド追加
- [ ] `core_schema.go`に全エンティティ種別のネスト定義を追加
- **完了条件**: `CoreSchema()`が全エンティティ種別に対して正しい`NestingDefs`を返し、テストがパスすること

### Phase 2: パーサー変更
- [ ] `parseEntity`の拡張（ネスト定義の解析）
- [ ] `parseNestedEntity`関数の追加
- [ ] ID自動生成ロジック（未指定時）
- [ ] 所有権の自動設定
- **完了条件**: ネスト定義付きYAMLをパースし、全子エンティティが正しいownerIDとkindでフラットなエンティティリストとしてグラフに追加されること

### Phase 3: 参照解決
- [ ] `resolvePathReference`関数の追加
- [ ] `ResolveReferences`の拡張
- [ ] パス表記の検証ロジック
- **完了条件**: ID参照とパス参照の両方が正しく解決され、存在しない参照はエラーとして検出されること

### Phase 4: シリアライザー変更
- [ ] `buildDocument`の拡張（ネスト出力）
- [ ] `buildEntityWithChildren`関数の追加
- [ ] フラット→ネスト変換ロジック
- **完了条件**: フラットなグラフからネスト構造のYAMLが生成され、ネストされたエンティティには`owner`フィールドが含まれないこと

### Phase 5: 検証ルール
- [ ] ネスト定義の整合性チェック
- [ ] IDのスラッシュ禁止チェック
- [ ] 親子関係の検証
- **完了条件**: 不正なネスト定義（不正な親子関係、IDにスラッシュ含む等）がエラーとして検出されること

### Phase 6: テスト
- [ ] ネスト定義のパーステスト
- [ ] フラット→ネスト変換テスト
- [ ] 参照解決テスト
- [ ] 統合テスト
- **完了条件**: 全テストがパスし、`go test ./...`が成功すること

### Phase 7: 仕様書更新
- [ ] `spec/concrete/17-yaml-syntax.md`の更新
- [ ] `spec/concrete/14-entity-kinds.md`の更新
- **完了条件**: 仕様書にネスト定義の構文と使用例が記載され、実装と一致すること

## 注意事項

- ネスト定義はあくまでYAML構文上の省略記法
- 内部的にはフラットなエンティティリストとして格納
- ラウンドトリップ（パース→シリアライズ）で意味が変わらないこと
