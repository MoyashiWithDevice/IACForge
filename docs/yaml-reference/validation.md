# Validation

[← README](README.md)

---

## 必須フィールド

| Object | Required Fields |
|--------|-----------------|
| Entity | id, kind, name |
| Relation | id, type, participants |

## 識別子ルール

- IDはそのスコープ内でユニークである必要があります
- IDは記述的で安定していることが望ましいです
- IDは命名規則（kebab-case推奨）に従うことが望ましいです
- IDにスラッシュ（`/`）を含めることはできません

---

## 命名規則

| 対象 | パターン | 例 |
|------|----------|-----|
| ID | kebab-case (推奨) | `srv-proxmox-01` |
| Kind | 小文字・単数形 | `server`, `vm` |
| Relation Type | snake_case | `connects`, `depends_on` |
| Property | snake_case | `cpu`, `memory`, `storage` |

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

### Nesting制約

| Constraint | Description |
|------------|-------------|
| no-slash-in-id | IDにスラッシュを含めることはできません |
| valid-nesting-parent | 親子関係がスキーマのネスト定義と一致すること |

### Cardinality制約

| Type | Cardinality | Description |
|------|-------------|-------------|
| connects | N:N | 多対多接続 |
| hosts | 1:N | 1ホスト、複数ゲスト |
| depends_on | N:N | 多対多依存関係 |
| belongs_to | N:N | 複数メンバー、複数グループ |
| applies_to | N:N | 1 ACL、複数ターゲット |
| listens_on | N:1 | 複数ポート、1インターフェース |
