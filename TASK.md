# IACForge 実装計画

## プロジェクト概要

IACForgeは、インフラストラクチャをモデルとして定義し、様々な表現を生成するためのフレームワークです。インフラは知識であり、一度定義したモデルからすべての表現を生成することを目指します。

## 実装フェーズ

### フェーズ1: コアオブジェクトモデル
**目標**: EntityとRelationの基本的なデータ構造を実装

#### 完了条件
1. Entityの作成・参照・更新が可能
2. Relationの作成・参照・更新が可能
3. すべてのEntity Kinds（19種類）が定義済み
4. すべてのRelation Types（11種類）が定義済み
5. Ownershipツリーが正しく構築される
6. 参照解決が正しく動作する
7. ユニットテストがすべてパスする
8. 仕様書 `spec/01-entity.md`, `spec/02-relation.md`, `spec/03-object-model.md` の要件を満たす

#### タスク
1. **Entity基底クラスの実装**
   - 共通プロパティ（id, kind, name, owner, description, status, tags, labels, metadata）
   - Identity管理
   - Ownershipツリー構築

2. **Relation基底クラスの実装**
   - 共通プロパティ（id, type, participants, description, status, tags, labels, metadata）
   - Directionality（directed/symmetric/undirected）

3. **Graphクラスの実装**
   - Objectコレクション管理
   - 参照解決
   - 整合性チェック

4. **Entity Kinds定義の実装**
   - site, rack, server, interface, cable, power_distribution
   - network, vlan, switch, router, firewall, acl, acl_rule
   - vm, container, application, open_port
   - storage, volume, cluster, availability_zone

5. **Relation Types定義の実装**
   - connects, hosts, depends_on, belongs_to
   - replicates_to, backs_up, monitors, managed_by, mounted_on, applies_to, listens_on

### フェーズ2: スキーマと検証
**目標**: モデルの構造定義と検証ルールの実装

#### 完了条件
1. Core Schemaが定義済み
2. すべてのプロパティタイプが実装済み
3. Validation Engineが動作し、ルールを評価できる
4. 14個のCore Validation Rulesがすべて実装済み
5. Validation Profileが作成可能
6. 検証結果がFiding形式で出力される
7. ユニットテストがすべてパスする
8. 仕様書 `spec/concrete/16-core-schema.md`, `spec/concrete/18-validation-rules.md` の要件を満たす

#### タスク
1. **Schema定義の実装**
   - プロパティタイプ定義
   - Entity Kind定義
   - Relation Type定義
   - 制約条件

2. **Validation Engineの実装**
   - ルール評価エンジン
   - Finding構造
   - Severity levels（info, warning, error）

3. **Core Validation Rulesの実装**
   - グラフ整合性ルール
   - Entityルール
   - Relationルール
   - Ownershipルール
   - Referenceルール

4. **Validation Profileの実装**
   - プロファイル構造
   - カスタムルール定義

### フェーズ3: YAML構文パーサー
**目標**: YAML形式でのモデル入出力

#### 完了条件
1. YAMLファイルからGraphをロード可能
2. GraphからYAMLファイルをエクスポート可能
3. EntityとRelationの構文が正しくパースされる
4. 参照が正しく解決される
5. 所有権（owner）が正しくパースされる
6. Round-trip変換がセマンティック等価性を保持する
7. サンプルYAMLファイルが正しく読み込める
8. ユニットテストがすべてパスする
9. 仕様書 `spec/concrete/17-yaml-syntax.md` の要件を満たす

#### タスク
1. **YAML Parserの実装**
   - Document structure解析
   - Entity構文解析
   - Relation構文解析
   - 参照解決

2. **YAML Serializerの実装**
   - Entity出力
   - Relation出力
   - フォーマット保持

3. **Round-trip対応**
   - セマンティック等価性保持
   - コメント preservation

### フェーズ4: クエリモデル
**目標**: グラフ上的な選択・走査機能

#### 完了条件
1. Select句でEntity/Relationを選択可能
2. Where句でフィルタリングが可能
3. Traverse句でグラフ走査が可能
4. Project句で結果投影が可能
5. すべての比較演算子（eq, ne, in, gt, lt等）が動作する
6. 論理演算子（and, or, not）が動作する
7. Ownership走査とRelation走査が動作する
8. クエリの組み合わせ（Composition）が可能
9. ユニットテストがすべてパスする
10. 仕様書 `spec/concrete/19-query-model.md` の要件を満たす

#### タスク
1. **Query Engineの実装**
   - Select句
   - Where句
   - Traverse句
   - Project句

2. **クエリオペレーターの実装**
   - 比較演算子
   - 論理演算子
   - 文字列演算子

3. **グラフ走査の実装**
   - Ownership走査
   - Relation走査
   - 深さ制御

### フェーズ5: プロジェクションモデル
**目標**: グラフ変換機能

#### 完了条件
1. Projection Engineがグラフ変換を実行可能
2. すべてのProjection Operations（10種類）が実装済み
3. Derived Objectsが正しく生成される
4. Provenance情報が正しく記録される
5. 複数のProjectionをチェーン可能
6. Projectionは副作用なし（元グラフを変更しない）
7. 同じ入力に対して同じ出力が得られる（決定性）
8. ユニットテストがすべてパスする
9. 仕様書 `spec/concrete/20-projection-model.md` の要件を満たす

#### タスク
1. **Projection Engineの実装**
   - Input clause
   - Operations
   - Output clause

2. **Projection Operationsの実装**
   - select, filter, traverse
   - aggregate, expand, annotate
   - group, flatten, enrich, transform

3. **Derived Objects管理**
   - Provenance追跡
   - ID生成

### フェーズ6: ビューとレンダリング
**目標**: モデルの可視化と出力

#### 完了条件
1. View定義が可能
2. 可視性ルールが動作する
3. グルーピングが可能
4. アノテーションが可能
5. 少なくとも1つのRendererが動作する（SVGまたはMarkdown）
6. Artifactが正しく生成される
7. Themeが適用可能
8. Layout Engineが動作する
9. ユニットテストがすべてパスする
10. 仕様書 `spec/concrete/21-rendering.md` の要件を満たす

#### タスク
1. **View定義の実装**
   - View構造
   - 可視性ルール
   - グルーピング
   - アノテーション

2. **Renderer基底クラスの実装**
   - Artifact生成
   - テーマサポート

3. **Concrete Renderersの実装**
   - SVG Renderer
   - Mermaid Renderer
   - Markdown Renderer
   - JSON Renderer

4. **Layout Engineの実装**
   - hierarchical layout
   - force-directed layout

### フェーズ7: 拡張モデル
**目標**: プラグイン可能な拡張システム

#### 完了条件
1. Extension Managerが拡張をロード可能
2. Namespace管理が動作する
3. 依存関係解決が可能
4. Entity Kinds拡張が可能
5. Relation Types拡張が可能
6. Validation Rules拡張が可能
7. Renderer拡張が可能
8. 拡張はコアモデルの整合性を破壊しない
9. ユニットテストがすべてパスする
10. 仕様書 `spec/13-extention-model.md` の要件を満たす

#### タスク
1. **Extension Managerの実装**
   - 拡張ロード
   - Namespace管理
   - 依存関係解決

2. **Extension Pointsの実装**
   - Entity Kinds拡張
   - Relation Types拡張
   - Validation Rules拡張
   - Renderer拡張

## 実装優先順位

1. フェーズ1（コアオブジェクトモデル）
2. フェーズ2（スキーマと検証）
3. フェーズ3（YAML構文パーサー）
4. フェーズ4（クエリモデル）
5. フェーズ5（プロジェクトションモデル）
6. フェーズ6（ビューとレンダリング）
7. フェーズ7（拡張モデル）

## 技術スタック

- **言語**: Go
- **YAML Parser**: `gopkg.in/yaml.v3`
- **テストフレームワーク**: Go標準テスト
- **ビルドツール**: Go modules

## 開発環境

```bash
# 依存関係インストール
go mod tidy

# テスト実行
go test ./...

# ビルド
go build ./...
```

## 参考仕様書

- `spec/00-philosophy.md` - プロジェクト哲学
- `spec/01-entity.md` - Entity定義
- `spec/02-relation.md` - Relation定義
- `spec/03-object-model.md` - オブジェクトモデル
- `spec/concrete/14-entity-kinds.md` - 具体的なEntity Kinds
- `spec/concrete/15-relation-types.md` - 具体的なRelation Types
- `spec/concrete/16-core-schema.md` - コアスキーマ
- `spec/concrete/17-yaml-syntax.md` - YAML構文
- `spec/concrete/18-validation-rules.md` - 検証ルール
- `spec/concrete/19-query-model.md` - クエリモデル
- `spec/concrete/20-projection-model.md` - プロジェクションモデル
- `spec/concrete/21-rendering.md` - レンダリング