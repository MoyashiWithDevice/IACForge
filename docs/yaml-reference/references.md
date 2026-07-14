# References

[← README](README.md)

---

## シンプル参照

```yaml
source: srv-proxmox-01
target: vm-web-01
```

## 修飾参照（フルパス）

```yaml
source: /site-tokyo-01/rack-a01/srv-proxmox-01
target: vm-web-01
```

## インターフェース参照

インターフェースはパス表記で参照します（`entity/interface`）：

```yaml
participants:
  - srv-proxmox-01/eno1
  - sw-core-01/port1
```

## 参照ルール

- Referencesは既存のObjectsを指す必要があります
- Unknown referenceは検証エラーとなります
- Interface referenceはパス表記を使用します
