# 仕様と実装の不一致修正計画

## 対象外
- 仕様外の追加実装（`no-slash-in-id`, `valid-nesting-parent`）→ 想定内

---

## Phase 1: 実装の修正（src/） ✅ 完了

### 1-1. Entity Properties の YAML struct tag 修正 ✅
- **ファイル:** `src/core/entity.go:72`
- **修正:** `yaml:"properties,omitempty"` → `yaml:"spec,omitempty"`

### 1-2. `single-owner` ルールの修正 ✅
- **ファイル:** `src/validation/engine.go:376-396`
- **修正:** ルート数チェック＋owner存在チェックに変更

### 1-3. `invalid-path` ルールの新規実装 ✅
- **ファイル:** `src/validation/engine.go:925-980`
- **内容:** パスセグメント存在チェック＋所有関係チェック

### 1-4. `ownership-tree` ルールの副作用修正 ✅ (追加修正)
- **ファイル:** `src/validation/engine.go:762-804`
- **問題:** `BuildOwnershipPaths()` 呼び出しが全Entityのパスを上書きしていた
- **修正:** ツリー接続性チェックに変更（副作用なし）

---

## Phase 2: 仕様の修正（spec/） ✅ 完了

### 2-1. spec 16: Entity Kind 定義の不足を補完 ✅
- **ファイル:** `spec/concrete/16-core-schema.md:123-143`
- **修正:** `power_distribution`, `vlan` を追加、19→21種に更新

### 2-2. spec 14: `interface` Entity Kind 定義の追加 ✅
- **ファイル:** `spec/concrete/14-entity-kinds.md:227`
- **修正:** `cable` の後に `interface` の定義セクションを追加

### 2-3. spec 18: ルール数の確認・修正 ✅
- **ファイル:** `src/validation/engine.go:128`
- **修正:** コメント "14 core validation rules" → "core validation rules"（数をハードコードしない）

---

## Phase 3: テスト ✅ 完了

- `single-owner` テスト: 既存テストでカバー
- `invalid-path` テスト: 3件追加済み（非存在エンティティ、所有関係エラー、正常パス）
- `ownership-tree` 副作用修正: 全31テスト通過

---

## Phase 4: ビルド・検証 ✅ 完了

```bash
go build ./...   ✅
go test ./...    ✅ (全パッケージ通過)
go vet ./...     ✅
```

---

## 実施した修正一覧

| Phase | ファイル | 内容 |
|-------|----------|------|
| 1-1 | `src/core/entity.go` | struct tag `properties` → `spec` |
| 1-2 | `src/validation/engine.go` | single-owner ルール再実装 |
| 1-3 | `src/validation/engine.go` | invalid-path ルール追加実装 |
| 1-4 | `src/validation/engine.go` | ownership-tree 副作用除去 |
| 2-1 | `spec/concrete/16-core-schema.md` | power_distribution, vlan 追加 |
| 2-2 | `spec/concrete/14-entity-kinds.md` | interface 定義セクション追加 |
| 2-3 | `src/validation/engine.go` | コメント更新 |
| 3 | `src/validation/engine_test.go` | invalid-path テスト3件追加 |
