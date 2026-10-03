# Recurring Calendar Expander

展开周期活动的本地日历清单，处理跨月日期及临时增减的实例。仅使用 Go 标准库，
只处理公历 `YYYY-MM-DD` 日期，不涉及时间或时区。

## 运行

```
go run ./cmd/calendar 输入.json --out 输出目录
```

生成两个文件：

- `instances.csv`：与查询窗口相交的实例，表头 `activity_id,start,end`，
  按开始日再按活动编号排序，保留完整起止日期（左闭右开）。
- `summary.json`：查询窗口、实例总数，以及各活动命中实例数（零命中也保留）。

任何非法定义都会整批失败：打印错误、退出码非零，且不生成结果文件。

## 输入格式

```json
{
  "window": { "start": "2026-02-01", "end": "2026-04-01" },
  "activities": [
    {
      "id": "gym",
      "start": "2026-01-28",
      "duration_days": 1,
      "recurrence": {
        "freq": "weekly",
        "interval": 1,
        "count": 8,
        "weekdays": ["WE", "FR"]
      },
      "add": ["2026-03-20"],
      "exclude": ["2026-02-13"]
    }
  ]
}
```

- `window`：查询窗口，`start`/`end` 均为 `YYYY-MM-DD`，左闭右开，须有正长度。
- `activities[]`：
  - `id`：活动编号，全局唯一。
  - `start`：起始日期。
  - `duration_days`：正整数持续天数；实例区间为 `[start, start+duration_days)`。
  - `recurrence`：可省略（不重复，仅起始日一次）；否则：
    - `freq`：`daily` | `weekly` | `monthly`。
    - `interval`：正整数间隔。
    - 终止条件二选一（必须且只能一个）：
      - `count`：正整数次数，统计有效规则日期并包含起始日。
      - `until`：包含该日的截止日期，不得早于起始日。
    - `weekdays`：仅 `weekly` 需要；不重复的星期集合
      （`MO,TU,WE,TH,FR,SA,SU`），且必须包含起始日星期。
  - `add`：额外日期列表，合并后去重，不受规则次数或截止日限制。
  - `exclude`：排除日期列表，优先于规则与额外日期，不补足次数；
    排除不存在的日期无副作用。

## 展开规则

- **daily**：从起始日起每隔 `interval` 天发生。
- **weekly**：以起始日所在周的周一为第零周，每隔 `interval` 周，
  枚举 `weekdays` 集合中的日期，跳过起始日之前的日期。
- **monthly**：保留起始日的日号，每隔 `interval` 个月；
  不存在该日的月份跳过且不消耗次数。

规则展开后合并 `add` 并去重，再扣除 `exclude`；同一活动不同开始日的实例不合并。

## 测试

```
go test ./...
```

示例输入见 `examples/sample.json`。
