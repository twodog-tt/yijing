# 起卦规则说明（MVP）

## 方法

采用三枚硬币法模拟：

| 点数 | 爻性 | 动爻 |
|------|------|------|
| 6 | 老阴 | 是 |
| 7 | 少阳 | 否 |
| 8 | 少阴 | 否 |
| 9 | 老阳 | 是 |

## 卦象

- **本卦**：六爻阴阳排列（自下而上）
- **动爻**：值为 6 或 9 的爻位
- **变卦**：动爻阴阳翻转后的卦象

## 数据存储

`divination_record.line_snapshot` 保存六爻 JSON 快照，便于结果页直接渲染，无需再查 `divination_line`。

示例：

```json
[
  {"position": 1, "value": 7, "is_yang": 1, "is_moving": 0},
  {"position": 6, "value": 6, "is_yang": 0, "is_moving": 1}
]
```
