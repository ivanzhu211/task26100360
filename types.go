// Package millsim 提供三轴铣削加工仿真：读取毛坯、刀具与刀路，
// 计算每一步的去除材料体积与剩余体积。单位均为毫米，Z 轴向上。
package millsim

// Vec3 是三维点/向量（毫米）。
type Vec3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// StockType 毛坯类型：长方体或竖直圆柱。
type StockType string

const (
	StockBox      StockType = "box"
	StockCylinder StockType = "cylinder"
)

// Stock 毛坯定义。Box 用 Min/Max；Cylinder 用 Center/Radius/ZMin/ZMax。
type Stock struct {
	Type   StockType  `json:"type"`
	Min    Vec3       `json:"min,omitempty"`
	Max    Vec3       `json:"max,omitempty"`
	Center [2]float64 `json:"center,omitempty"`
	Radius float64    `json:"radius,omitempty"`
	ZMin   float64    `json:"zmin,omitempty"`
	ZMax   float64    `json:"zmax,omitempty"`
}

// ToolType 刀具类型：平底立铣刀或球头刀。
type ToolType string

const (
	ToolFlat ToolType = "flat"
	ToolBall ToolType = "ball"
)

// Tool 刀具。刀位点为刀尖最低点；刃长以上不参与切削。
type Tool struct {
	ID          string   `json:"id"`
	Type        ToolType `json:"type"`
	Diameter    float64  `json:"diameter"`
	FluteLength float64  `json:"fluteLength"`
}

// MoveType 移动类型。
type MoveType string

const (
	MoveRapid  MoveType = "rapid"  // 快速直线
	MoveLinear MoveType = "linear" // 切削直线
	MoveArc    MoveType = "arc"    // XY 平面圆弧（可带 Z 变化，即螺旋）
)

// ArcDirection 圆弧方向（从 +Z 向下看）。
type ArcDirection string

const (
	ArcCW  ArcDirection = "cw"
	ArcCCW ArcDirection = "ccw"
)

// Move 一步中的一个移动。Arc 需要 Center 与 Direction；
// 起点与终点重合的圆弧按整圆处理。
type Move struct {
	Type      MoveType     `json:"type"`
	To        Vec3         `json:"to"`
	Center    [2]float64   `json:"center,omitempty"`
	Direction ArcDirection `json:"direction,omitempty"`
}

// Step 程序中的一步：一把刀 + 起点 + 有序移动序列。
type Step struct {
	ToolID string `json:"toolId"`
	Start  Vec3   `json:"start"`
	Moves  []Move `json:"moves"`
}

// Job 仿真输入：毛坯、刀具表、程序步骤与可选网格分辨率。
type Job struct {
	Stock      Stock   `json:"stock"`
	Tools      []Tool  `json:"tools"`
	Steps      []Step  `json:"steps"`
	Resolution float64 `json:"resolution,omitempty"` // dexel 网格边长，默认 0.2mm
}

// StepResult 单步仿真结果。
type StepResult struct {
	Index           int     `json:"index"`
	RemovedVolume   float64 `json:"removedVolume"`
	RemainingVolume float64 `json:"remainingVolume"`
}

// Result 仿真输出。
type Result struct {
	InitialVolume float64      `json:"initialVolume"`
	Steps         []StepResult `json:"steps"`
	FinalVolume   float64      `json:"finalVolume"`
}
