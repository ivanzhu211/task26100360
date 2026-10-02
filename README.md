# millsim — 三轴铣削加工仿真

读取毛坯、刀具与刀路（本地 JSON 文件），在电脑上把整份程序走一遍，
算出每一步实际去掉的材料体积和剩余体积。单位：毫米，Z 轴向上，刀具始终竖直。

## 构建与测试

```sh
go build ./...
go test ./...
```

## 命令行用法

```sh
millsim -job program.json -out result.json
```

## 输入 JSON 格式

```json
{
  "stock": {"type": "box", "min": {"x":0,"y":0,"z":0}, "max": {"x":100,"y":60,"z":30}},
  "tools": [{"id":"T1","type":"flat","diameter":10,"fluteLength":25}],
  "steps": [
    {
      "toolId": "T1",
      "start": {"x":10,"y":15,"z":40},
      "moves": [
        {"type":"rapid","to":{"x":10,"y":15,"z":26}},
        {"type":"linear","to":{"x":50,"y":15,"z":26}},
        {"type":"arc","center":[50,25],"direction":"ccw","to":{"x":50,"y":35,"z":24}}
      ]
    }
  ],
  "resolution": 0.2
}
```

- `stock`：`box`（`min`/`max`，各边与坐标轴平行）或
  `cylinder`（`center`/`radius`/`zmin`/`zmax`，轴线竖直）。
- `tools`：`flat`（平底立铣刀）或 `ball`（球头刀，底部为同半径半球）。
  刀位点为刀尖最低点；从刀尖往上 `fluteLength` 为切削部分，刃长以上不切料。
- `steps[].moves`：`rapid` / `linear`（直线）与 `arc`（XY 平面圆弧，
  给出 `center` 与 `direction`=`cw`/`ccw`，终点可带 Z 变化即螺旋；
  起终点重合按整圆处理）。快速移动同样会去料。
- `resolution`：dexel 网格边长（mm），缺省 0.2。越小越精确、越慢。

## 输出 JSON 格式

```json
{
  "initialVolume": 150811.2,
  "steps": [
    {"index": 0, "removedVolume": 3644.34, "remainingVolume": 147166.86}
  ],
  "finalVolume": 145861.22
}
```

## 作为库使用

```go
job, _ := millsim.LoadJob("program.json")
res, _ := millsim.Run(job)
millsim.SaveResult("result.json", res)
```

## 实现说明

毛坯用 dexel（竖直柱）模型表示：XY 规则网格上每格保存若干材料区间。
刀具沿刀路按不超过半格的步距离散放置，每个位置把切削部分覆盖的
格子区间减去。同一处材料只会被减去一次，因此重复走同一条路径去除量为 0。
体积精度随 `resolution` 提高；测试表明典型开槽/螺旋下刀与解析值误差 < 2%。
