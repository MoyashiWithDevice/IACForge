# Validation

[← README](README.md)

---

## 必須フィールド

| Object | Required Fields |
|--------|-----------------|
| Entity | id, kind, name |
| Relation | id, type, participants |

## Ownership検証

- Root Entityはownerを指定してはなりません
- Non-root Entityは正確に1つのownerを指定する必要があります
- Owner識別子は既存のEntityを参照する必要があります
- Ownershipは正確に1つのツリーを形成する必要があります
- Ownershipはサイクルを含んではなりません

## Reference検証

- Referencesは既存のObjectsを指す必要があります
- Interface referenceはパス表記を使用します（`entity/interface`）
- Unknown referenceは検証エラーとなります

## 識別子ルール

- IDはそのスコープ内でユニークである必要があります
- IDは記述的で安定していることが望ましいです
- IDは命名規則（kebab-case推奨）に従うことが望ましいです

---

## 命名規則

### 推奨パターン

| Pattern | Example |
|---------|---------|
| kebab-case | `srv-proxmox-01` |
| snake_case | `mgmt_network_01` |

### Kind命名

- 小文字を使用
- 単数形を使用
- 例: `server`, `vm`, `interface`

### Relation Type命名

- snake_caseを使用
- 動詞または動詞句を使用
- 例: `connects`, `hosts`, `depends_on`

### Property命名

- snake_caseを使用
- 例: `cpu_cores`, `memory_gb`, `ip_address`

---

## Graph制約

### Ownership制約

| Constraint | Description |
|------------|-------------|
| single_owner | Root以外のEntityは正確に1つのownerを持つ |
| tree_structure | Ownershipは正確に1つのツリーを形成 |
| no_cycles | Ownershipはサイクルを含まない |
| owner_exists | Owner識別子は既存のEntityを参照する |

### Reference制約

| Constraint | Description |
|------------|-------------|
| valid_reference | Referencesは既存のObjectsを指す |
| unique_id | Object識別子はユニークである |
| relation_exists | Relationsは既存のObjectsを参照する |

### Cardinality制約

| Type | Cardinality | Description |
|------|-------------|-------------|
| connects | N:N | 多対多接続 |
| hosts | 1:N | 1ホスト、複数ゲスト |
| depends_on | N:N | 多対多依存関係 |
| belongs_to | N:N | 複数メンバー、複数グループ |
| applies_to | N:N | 1 ACL、複数ターゲット |
| listens_on | N:1 | 複数ポート、1インターフェース |
