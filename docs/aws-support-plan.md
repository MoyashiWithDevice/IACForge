# AWS対応 実装計画

## 前提: 現状の重要事実

- `src/extension/` の拡張システム（4拡張ポイント）はコード完成済みだが **実行時パスに未配線**（`session.go:48` は `schema.CoreSchema()` 固定、`main.go` は `LoadFromDir` 未使用、MCPに拡張ロードツールなし）
- 仕様14の `aws.vpc` 命名規則は拡張を前提。よって **AWS拡張を動かすにはまず拡張ローダーを組み込む必要がある**
- コア関係タイプの参加者制約が core kind 固定（`src/core/types/types.go`、`core_schema.go`）→ aws kind を参加させると `valid-participant-kind` 警告が出る
- `validate_graph` にプロパティ型/列挙値の検証ルールがない
- 単一ルート所有権ツリー強制（`ruleSingleOwner`/`ruleRootEntity`）がAWSの複数アカウント表現を阻害

---

## Phase 0: 拡張システムの実行時配線（AWS対応の前提）

### 0-1. MCPセッションへの拡張統合 — `src/mcp/session.go`

- `SessionData` に `Extensions *extension.Manager` を追加
- `GetOrCreate` で以下を構築:
  1. `schema.CoreSchema()` 生成
  2. `validation.NewEngine(s)` + `RegisterCoreRules`
  3. `extension.NewManager()` → 4つの拡張ポイントを登録（EntityKinds / RelationTypes / ValidationRules / Renderers）※各拡張ポイントは `s` と `v` へのポインタを保持するため、後のツールから拡張kindが見える
  4. ビルトインAWS拡張（Phase 3）を `Register` + `LoadAll`
  5. 環境変数 `IACFORGE_EXTENSIONS` 指定時は `LoadFromDir` で `.so` プラグインもロード

### 0-2. CLIへの拡張統合 — `cmd/iacforge/main.go`

- `cmdValidate`/`cmdInfo`/`cmdRender`/`cmdQuery` で使用する schema/validation を、拡張マネージャ経由で構築する共通ヘルパー（例: `newSchemaWithExtensions()`）に置換
- `validate` コマンドに `--extensions <dir>` フラグ追加

### 0-3. 拡張ロード用MCPツール — `src/mcp/tools_extension.go`（新規）

- `register_extension`（プラグイン管理）
- `load_extension_dir`（`.so` ディレクトリロード）
- `list_extensions` / `list_extension_kinds`
- `server.go` の `NewMCPServer` に登録

---

## Phase 1: AWS拡張のための基盤強化

### 1-1. リレーション参加者制約の拡張

- `src/extension/types.go`: `RelationTypeContribution` に `Augment bool` フィールド追加（既存core typeへの**追加的**参加者拡張）
- `src/extension/relation_types.go`: `Register` で `Augment==true` のとき、既存定義の SourceKinds/TargetKinds に contributions の kind を**結合**（`validateNoCoreConflict` の衝突判定をバイパス）
- これにより `belongs_to` / `depends_on` / `hosts` / `monitors` / `backs_up` に aws kind を追加可能に

### 1-2. プロパティスキーマ検証ルールの追加 — `src/validation/engine.go`

- 新コアルール `valid-property`（Severity=Warning）: 各エンティティ/リレーションの `spec` プロパティを `schema.ValidateProperty`（`schema.go:210`）で検証（型・enum・required・min/max）。未定義プロパティも警告
- Warningにすることで既存モデルを壊さず、拡張kindのプロパティ誤りを検出可能に

### 1-3. 複数アカウント対応 — `src/validation/engine.go`

- `ruleRootEntity`/`ruleSingleOwner`/`ruleOwnershipTree` に、`aws.organization` のような「ルート権限を持つkind」を許可する拡張フックを追加（`kinds.AllKinds` とは別に、ルートとして許可するkindリスト）。AWS拡張が `aws.organization` をルートとして登録

---

## Phase 2: AWS拡張 kind / relation 定義（網羅）

namespace = `aws`、ID/kindは小文字単数。所有権ツリーの基本形:

```
aws.organization
└── aws.account
    ├── aws.region
    │   ├── aws.availability_zone
    │   ├── aws.vpc
    │   │   ├── aws.subnet
    │   │   │   ├── aws.ec2 / aws.rds / aws.load_balancer / aws.nat_gateway
    │   │   ├── aws.security_group / aws.route_table / aws.internet_gateway / aws.network_acl
    │   └── aws.s3_bucket / aws.elasticache / aws.ebs_volume (AZ直下)
    ├── aws.iam_* / aws.lambda / aws.sqs_queue / aws.sns_topic / aws.dynamodb_table / aws.api_gateway
    └── aws.route53_zone / aws.cloudfront_distribution / aws.cloudwatch_*
```

### kind一覧（約46種）と主なプロパティ

| グループ | kind | 主なプロパティ |
|---|---|---|
| 組織/ID | `aws.organization` | name, org_root, org_units |
| | `aws.account` | account_id, alias, email, org_path |
| IAM | `aws.iam_user` `iam_group` `iam_role` `iam_policy` `iam_instance_profile` | arn, path, assume_role_policy, policy_document |
| ネットワーク | `aws.region` | region_code, partition, opt_in_status |
| | `aws.availability_zone` | zone_name, zone_id, group_name |
| | `aws.vpc` | cidr_block, tenancy, dns_support, flow_logs |
| | `aws.subnet` | cidr_block, map_public_ip, default_for_az |
| | `aws.route_table` / `aws.route` | destination_cidr, gateway_id, nat_gateway_id |
| | `aws.internet_gateway` `aws.nat_gateway` `aws.vpc_peering_connection` `aws.transit_gateway` `aws.network_acl` | connectivity_type, auto_accept |
| | `aws.security_group` / `aws.security_group_rule` | from_port, to_port, protocol, source_security_group(ref) |
| | `aws.elastic_ip` | domain, public_ip, association |
| コンピュート | `aws.ec2` | instance_type, ami(ref), key_pair(ref), subnet(ref), security_groups(ref), user_data, iam_instance_profile(ref) |
| | `aws.ami` `aws.key_pair` `aws.launch_template` | image_id, key_type |
| | `aws.auto_scaling_group` | min/max/desired, launch_template(ref), vpc_zone_identifier |
| | `aws.ebs_volume` `aws.ebs_snapshot` | volume_type, iops, encrypted, kms_key_id |
| | `aws.lambda_function` | runtime, handler, role(ref), memory_size, vpc_config |
| ストレージ | `aws.s3_bucket` `aws.efs` | versioning, encryption, lifecycle_rules |
| DB | `aws.rds` `aws.dynamodb_table` `aws.elasticache` | engine, instance_class, multi_az, billing_mode |
| LB | `aws.load_balancer` `aws.target_group` `aws.listener` | type(alb/nlb), scheme, protocol, port |
| 統合 | `aws.sqs_queue` `aws.sns_topic` `aws.api_gateway` `aws.cloudfront_distribution` `aws.eventbridge_rule` | fifo, visibility_timeout, origin, event_pattern |
| 監視 | `aws.cloudwatch_alarm` `aws.cloudwatch_log_group` `aws.cloudwatch_dashboard` | metric_name, threshold, retention_days |
| DNS | `aws.route53_zone` `aws.route53_record` | private, record_name, type, ttl, alias_target |

### リレーション（aws拡張で新規定義 + coreへの参加者追加）

- 新規: `aws.associates`（SG→VPC等）、`aws.attaches`（EBS→EC2）、`aws.launches`（ASG→EC2）、`aws.routes`（RT→route）、`aws.serves`（LB→TG）、`aws.forwards`（listener→TG）、`aws.triggers`（EventBridge→Lambda）、`aws.subscribes`（SNS→SQS/Lambda）、`aws.invokes`（API GW→Lambda）、`aws.grants`（Policy→IAM主体）、`aws.assumes`（Role→Role）
- coreへの参加者追加（`Augment`）: `belongs_to`（subnet→vpc、ec2→subnet等）、`depends_on`、`hosts`（ec2→application）、`monitors`（cloudwatch_alarm→ec2）、`backs_up`（rds→snapshot）

---

## Phase 3: ビルトインAWS拡張の実装

- 新ディレクトリ `src/extension/builtin/aws/`（`aws.go` + `aws_test.go`）: `func Extension() *extension.Extension` を返すGoパッケージ。`.so` ではなく**コード内ビルトイン**として登録（Goプラグインのバージョン縛りを回避、即利用可能）。外部向け `.so` ビルドは `docs/extension-reference.md` の既存フローで任意対応
- 1つの `schema.EntityKindDefinition` 群をPhase 2の定義に従って実装。enum値・min/max制約・nesting定義（自動 `belongs_to`）・relation participant制約をフル設定
- `extension.Manifest`: `ID: "iacforge.aws"`, `Namespace: "aws"`

---

## Phase 4: テスト

- **単体**: 拡張配線（session/CLIでaws kindが `list_entity_kinds` に現れる）、`Augment` によるparticipant拡張、`valid-property` ルール、AWS拡張の各kind定義の整合
- **統合**: サンプルAWSモデルYAML（`testdata/aws-example.yaml`、account→region→vpc→subnet→ec2+s3+rds+lambda+sqs を網羅）の `load → validate → serialize` round-trip、MCPの `add_entity`/`query`/`validate_graph` でaws kindが全ツールで機能
- `go test ./...` 全件成功、カバレッジ80%目標（AGENTS.md準拠）

---

## Phase 5: ドキュメント

- 新規 `spec/concrete/22-aws-extension.md`: aws kind全定義・リレーション・所有権ツリー・ID/ARN規約
- `spec/concrete/14-entity-kinds.md` 末尾のVendor Kinds節と整合を確認
- `docs/yaml-reference/` にAWSセクション追加（`entity-kinds.md` 等）

---

## リスク・設計判断

1. **単一ルート制約**: `aws.organization` をルート権限kindにすることで複数アカウント対応（1-3）。従来のオンプレモデル（regionがルート）も共存可
2. **`aws.region` vs core `region`**: 直近コミットでcore regionはAWS寄りに改名済みだが、core kindへは拡張からプロパティ追加不可のため、AWS完全自己完結のため `aws.region` を新設。core `region` は汎用のまま温存
3. **ARN表現**: entity `id` には短名、`labels.arn` にARNを格納する規約（`@` 参照は既存ID解決のまま）
4. **`valid-property` をWarningに**: 既存モデルの未知プロパティ許容を壊さないための慎重措置。導入後にエラー昇格は別途判断
5. **コミット規模**: Phase 0→1は基盤変更（既存テストへの影響あり）。Phase 2/3は追加のみで安全

---

## フェーズ別タスクブレークダウン

### Phase 0: 拡張システムの実行時配線

| タスク | 内容 | 対象ファイル |
|--------|------|-------------|
| T0-1 | `SessionData` に `Extensions *extension.Manager` を追加。`GetOrCreate` で schema/validation/マネージャ構築、4拡張ポイント登録、ビルトイン拡張 `LoadAll`、`IACFORGE_EXTENSIONS` 指定時 `LoadFromDir` | `src/mcp/session.go` |
| T0-2 | 共通ヘルパー `newSchemaWithExtensions()` を新設し `validate`/`info`/`render`/`query` の schema/validation 構築を置換。`validate --extensions <dir>` フラグ追加 | `cmd/iacforge/main.go` |
| T0-3 | `load_extension_dir` / `list_extensions` / `list_extension_kinds` MCPツールを実装し `NewMCPServer` に登録 | `src/mcp/tools_extension.go`（新規）、`src/mcp/server.go` |
| T0-4 | ビルトイン拡張登録の共通化（session/CLIで同一経路のヘルパーを共有） | `src/extension/`、`src/mcp/session.go` |

### Phase 1: 基盤強化

| タスク | 内容 | 対象ファイル |
|--------|------|-------------|
| T1-1 | `RelationTypeContribution` に `Augment bool` を追加し、`Augment==true` 時に既存core typeへ参加者kindを結合（衝突判定バイパス） | `src/extension/types.go`、`src/extension/relation_types.go` |
| T1-2 | コアルール `valid-property`（Warning）を追加。`schema.ValidateProperty` を利用し型・enum・required・min/max・未知プロパティを検証 | `src/validation/engine.go` |
| T1-3 | ルート権限kindの拡張フックを追加。`aws.organization` をルートとして許可 | `src/validation/engine.go` |

### Phase 2: AWS kind/relation 定義

| タスク | 内容 | 対象ファイル |
|--------|------|-------------|
| T2-1 | 組織/ID/IAM kind: `aws.organization` `aws.account` `aws.iam_user` `aws.iam_group` `aws.iam_role` `aws.iam_policy` `aws.iam_instance_profile`（7種） | `src/extension/builtin/aws/` |
| T2-2 | ネットワーク kind: `aws.region` `aws.availability_zone` `aws.vpc` `aws.subnet` `aws.route_table` `aws.route` `aws.internet_gateway` `aws.nat_gateway` `aws.security_group` `aws.security_group_rule` `aws.elastic_ip` `aws.network_acl`（12種） | `src/extension/builtin/aws/` |
| T2-3 | コンピュート kind: `aws.ec2` `aws.ami` `aws.key_pair` `aws.launch_template` `aws.auto_scaling_group` `aws.ebs_volume` `aws.ebs_snapshot` `aws.lambda_function`（8種） | `src/extension/builtin/aws/` |
| T2-4 | ストレージ/DB kind: `aws.s3_bucket` `aws.efs` `aws.rds` `aws.dynamodb_table` `aws.elasticache`（5種） | `src/extension/builtin/aws/` |
| T2-5 | LB/統合/監視/DNS kind: `aws.load_balancer` `aws.target_group` `aws.listener` `aws.sqs_queue` `aws.sns_topic` `aws.api_gateway` `aws.cloudfront_distribution` `aws.eventbridge_rule` `aws.cloudwatch_alarm` `aws.cloudwatch_log_group` `aws.route53_zone` `aws.route53_record`（12種） | `src/extension/builtin/aws/` |
| T2-6 | aws拡張リレーション新規11種: `associates` `attaches` `launches` `routes` `serves` `forwards` `triggers` `subscribes` `invokes` `grants` `assumes` | `src/extension/builtin/aws/` |
| T2-7 | core relationへの参加者Augment定義: `belongs_to` `depends_on` `hosts` `monitors` `backs_up` `mounted_on` | `src/extension/builtin/aws/` |

### Phase 3: ビルトインAWS拡張

| タスク | 内容 | 対象ファイル |
|--------|------|-------------|
| T3-1 | パッケージ骨格 + `Manifest`（ID `iacforge.aws` / Namespace `aws` / `Extension()` エクスポート） | `src/extension/builtin/aws/aws.go` |
| T3-2 | Phase 2のkind定義・nesting・constraints・relation participant制約を実装 | `src/extension/builtin/aws/aws.go` |
| T3-3 | 動作確認: session/CLIで `list_entity_kinds` に `aws.*` が出現、サンプルモデルが `validate` をパス | 手動検証 + テスト |

### Phase 4: テスト

| タスク | 内容 | 対象ファイル |
|--------|------|-------------|
| T4-1 | Phase 0/1の単体テスト（拡張配線、Augment、valid-property、ルートフック） | `src/mcp/*_test.go`、`src/extension/*_test.go`、`src/validation/*_test.go` |
| T4-2 | AWS拡張の単体テスト（kind定義整合、enum/min-max、nesting、relation participant） | `src/extension/builtin/aws/aws_test.go` |
| T4-3 | 統合: `testdata/aws-example.yaml`（account→region→vpc→subnet→ec2+s3+rds+lambda+sqs を網羅）の load→validate→serialize round-trip | `testdata/aws-example.yaml`（新規）+ 統合テスト |
| T4-4 | MCP統合テスト（`add_entity` / `query` / `validate_graph` で aws kind が動作） | `src/mcp/*_test.go` |
| T4-5 | `go test ./...` 全件成功 + カバレッジ80%確認 | 全体 |

### Phase 5: ドキュメント

| タスク | 内容 | 対象ファイル |
|--------|------|-------------|
| T5-1 | AWS拡張仕様書: kind全定義・リレーション・所有権ツリー・ID/ARN規約 | `spec/concrete/22-aws-extension.md`（新規） |
| T5-2 | `spec/concrete/14-entity-kinds.md` 末尾のVendor Kinds節と整合確認 | `spec/concrete/14-entity-kinds.md` |
| T5-3 | `docs/yaml-reference/` にAWSセクション追加 | `docs/yaml-reference/entity-kinds.md` 等 |
