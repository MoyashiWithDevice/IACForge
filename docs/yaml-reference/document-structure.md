# Document Structure

[← README](README.md)

---

YAML documentは単一のGraphを表現します。

## 基本構造

```yaml
objects:
  # すべてのEntityとRelationをここに記述
```

## コメント

Commentsはround-trip変換時に保持されます。

```yaml
# Site情報
objects:
  - id: site-tokyo-01
    kind: site
    name: Tokyo Datacenter 1
    # プライマリロケーション
    status: active
```

## 記述順序

- Objectの順序には意味はありません
- Implementationは可能な限り順序を保持すべきです
