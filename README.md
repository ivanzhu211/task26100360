# millsim — 三轴铣削加工仿真

把整份加工程序在电脑上先走一遍，算出每一步实际去掉的材料体积和剩余体积。
包含一个 Go 库（`millsim` 包）和一个命令行工具（`cmd/millsim`），供上层 CAM
服务在本地调用；无网络、无数据库，输入输出都是本地 JSON 文件。

## 用法

```sh
go build ./cmd/millsim
./millsim <input.json> <output.json>
```

库调用：

```go
job := millsim.Job{ /* ... */ }
res, err := millsim.Simulate(job)
```

## 输入 JSON

```json
{
  "stock": { "type": "box", "origin": [0, 0, 0], "size": [100, 60, 25] },
  "tools": {
    "T1": { "type": "flat", "diameter": 10, "fluteLength": 30 },
    "T2": { "type": "ball", "diameter": 6, "fluteLength": 20 }
  },
  "resolution": 0.25,
  "steps": [
    {
      "name": "rough-slot",
      "tool": "T1",
      "start": [10, 30, 40],
      "moves": [
        { "type": "rapid", "to": [10, 30, 15] },
        { "type": "linear", "to": [90, 30, 15] },
        { "type": "arc", "center": [90, 20], "cw": true, "to": [100, 20, 12] }
      ]
    }
  ]
}
```

- 单位毫米，Z 轴向上，刀具始终竖直，刀位点是刀尖最低点。
- `stock.type`：`box`（`origin` 为最小角点，`size` 为三边长）或
  `cylinder`（`origin` 为底面圆心，`radius` + `height`，轴线竖直）。
- `tools.*.type`：`flat`（平底立铣刀）或 `ball`（球头刀，底部为同半径半球）。
  `fluteLength` 为刃长：从刀尖向上这一段切削，刃长以上不切料。
- `resolution`：仿真的网格边长（mm），缺省 0.5。越小越精确、越慢。
- 每步 `start` 为该步起点（须高于毛坯最高处），`moves` 按顺序执行：
  - `rapid` / `linear`：直线到 `to`。两者都会切料（规则如此）。
  - `arc`：XY 平面内绕 `center` 的圆弧，`cw` 区分顺/逆时针（从 +Z 看），
    `to` 可带 Z 变化（螺旋）。起点与终点 XY 重合时按整圆处理。

## 输出 JSON

```json
{
  "initialVolume": 150000,
  "steps": [
    { "name": "rough-slot", "removedVolume": 8790, "remainingVolume": 141210 }
  ],
  "finalVolume": 141210
}
```

`removedVolume` 只计本次新切掉的材料：已切过的区域和空行程不计入。

## 方法

毛坯用规则 XY 网格的柱高场（dexel/heightfield）表示，每列记录材料顶面 Z。
刀具沿轨迹扫掠，对覆盖到的每一列求刀底包络面高度并下切；平底刀底面为平面，
球头刀底面为半球面。圆弧按弓高误差不超过 1/4 网格边长细分为弦。刃长限制：
柱顶高出刀底包络面超过刃长时该列不切（高场无法表示悬空料，保守处理）。

体积误差主要来自网格离散，典型相对误差 < 0.5%（见测试日志）。初始体积、
各步去除量与剩余体积在同一离散模型内自洽：剩余 = 初始 − 累计去除。

## 测试

```sh
go test ./...
```

覆盖：平底刀直线开槽、整圆环槽、螺旋下刀、球头刀直槽（均与解析/独立数值
积分结果对比）、同路径重复走刀去除量为 0、圆柱毛坯初始体积。
