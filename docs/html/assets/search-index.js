/*
 * IACForge Reference - search index
 * Client-side search data. Kept dependency-free and static.
 * Each entry: { path, title, section, description, keywords }
 */

window.IACFORGE_SEARCH_INDEX = [
  {
    path: "index.html",
    title: "Reference",
    section: "Overview",
    description: "IACForgeのYAMLファイル作成のための完全版リファレンスのホーム。",
    keywords: "home overview top index reference yaml ホーム 概要 はじめに",
  },
  {
    path: "document-structure.html",
    title: "Document Structure",
    section: "YAML Reference",
    description: "YAMLドキュメントの基本構造、コメント、記述順序。",
    keywords: "document structure yaml graph objects comment order ドキュメント 構造 コメント 順序",
  },
  {
    path: "entity-syntax.html",
    title: "Entity Syntax",
    section: "YAML Reference",
    description: "Entity共通プロパティ、必須プロパティ、ステータス値、ネスト定義。",
    keywords: "entity syntax id kind name attributes owner status planned active maintenance deprecated offline standby tags labels extensions nested nest エンティティ 構文 必須 ステータス ネスト",
  },
  {
    path: "entity-kinds.html",
    title: "Entity Kinds",
    section: "YAML Reference",
    description: "全Entity種類の定義とプロパティ。site, rack, server, vm, network, storageなど。",
    keywords: "entity kinds site rack server interface cable power_distribution network vlan switch router firewall acl acl_rule vm container application open_port storage volume cluster availability_zone mode trunk access hybrid tagged kind 種類 定義",
  },
  {
    path: "relation-syntax.html",
    title: "Relation Syntax",
    section: "YAML Reference",
    description: "Relation共通構文、participantフォーマット（リスト・マップ形式）。",
    keywords: "relation syntax id type participants source target symmetric directed attributes リレーション 構文 参加者",
  },
  {
    path: "relation-types.html",
    title: "Relation Types",
    section: "YAML Reference",
    description: "全Relation種類の定義とプロパティ。connects, hosts, depends_on, belongs_toなど。",
    keywords: "relation types connects hosts depends_on belongs_to replicates_to backs_up monitors managed_by mounted_on applies_to listens_on 種類",
  },
  {
    path: "references.html",
    title: "References",
    section: "YAML Reference",
    description: "参照構文（シンプル、修飾、インターフェース、パス参照）と参照ルール。",
    keywords: "references reference path qualified interface simple nested participant source target 参照 パス 修飾",
  },
  {
    path: "validation.html",
    title: "Validation",
    section: "YAML Reference",
    description: "検証ルール、命名規則、Graph制約（Ownership, Reference, Nesting, Cardinality）。",
    keywords: "validation rules constraints ownership tree no_cycles unique_id cardinality naming kebab-case 検証 ルール 制約 命名規則 カーディナリティ",
  },
  {
    path: "example.html",
    title: "Complete Example",
    section: "YAML Reference",
    description: "完全なインフラモデルの例。sites, racks, servers, VMs, ACLs, relations。",
    keywords: "example complete full model yaml infrastructure sites racks servers vms acl relations 例 完全 サンプル",
  },
];
