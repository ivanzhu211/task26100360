package millsim

import (
	"math"
	"testing"
)

// 平底刀直线开槽：先垂直下刀，再直线切削。
// 去除体积 = 槽深 × 胶囊形 footprint（2R·L + πR²）。
func TestStraightSlotAnalytic(t *testing.T) {
	const (
		R     = 5.0  // 刀具半径
		depth = 4.0  // 槽深
		L     = 40.0 // 槽长
	)
	job := Job{
		Stock: Stock{Type: StockBox, Min: Vec3{0, 0, 0}, Max: Vec3{60, 30, 20}},
		Tools: []Tool{{ID: "T1", Type: ToolFlat, Diameter: 2 * R, FluteLength: 20}},
		Steps: []Step{{
			ToolID: "T1",
			Start:  Vec3{X: 10, Y: 15, Z: 30},
			Moves: []Move{
				{Type: MoveRapid, To: Vec3{X: 10, Y: 15, Z: 20 - depth}},
				{Type: MoveLinear, To: Vec3{X: 10 + L, Y: 15, Z: 20 - depth}},
			},
		}},
		Resolution: 0.1,
	}
	res, err := Run(job)
	if err != nil {
		t.Fatal(err)
	}
	want := depth * (2*R*L + math.Pi*R*R)
	got := res.Steps[0].RemovedVolume
	if e := math.Abs(got-want) / want; e > 0.015 {
		t.Fatalf("slot removed = %v, want %v (err %.2f%%)", got, want, e*100)
	}
	if rem := res.Steps[0].RemainingVolume; math.Abs(rem-(res.InitialVolume-got)) > 1e-6 {
		t.Fatalf("remaining %v inconsistent with removed %v", rem, got)
	}
}

// 螺旋下刀 + 最后一圈整圆平底：去除体积 = π(r+R)² × 深度。
func TestHelicalPlungeAnalytic(t *testing.T) {
	const (
		R     = 5.0 // 刀具半径
		r     = 3.0 // 螺旋半径
		depth = 6.0 // 最终刀尖深度（相对毛坯顶面）
	)
	center := [2]float64{30, 15}
	start := Vec3{X: center[0] + r, Y: center[1], Z: 30}
	job := Job{
		Stock: Stock{Type: StockBox, Min: Vec3{0, 0, 0}, Max: Vec3{60, 30, 20}},
		Tools: []Tool{{ID: "T1", Type: ToolFlat, Diameter: 2 * R, FluteLength: 20}},
		Steps: []Step{{
			ToolID: "T1",
			Start:  start,
			Moves: []Move{
				{Type: MoveRapid, To: Vec3{X: start.X, Y: start.Y, Z: 22}},
				// 三圈螺旋，每圈降 2mm：22 -> 16
				{Type: MoveArc, Center: center, Direction: ArcCW, To: Vec3{X: start.X, Y: start.Y, Z: 20}},
				{Type: MoveArc, Center: center, Direction: ArcCW, To: Vec3{X: start.X, Y: start.Y, Z: 18}},
				{Type: MoveArc, Center: center, Direction: ArcCW, To: Vec3{X: start.X, Y: start.Y, Z: 20 - depth}},
				// 最后一圈整圆平底，把螺旋坡底铣平
				{Type: MoveArc, Center: center, Direction: ArcCW, To: Vec3{X: start.X, Y: start.Y, Z: 20 - depth}},
			},
		}},
		Resolution: 0.1,
	}
	res, err := Run(job)
	if err != nil {
		t.Fatal(err)
	}
	want := math.Pi * (r + R) * (r + R) * depth
	got := res.Steps[0].RemovedVolume
	if e := math.Abs(got-want) / want; e > 0.015 {
		t.Fatalf("helical removed = %v, want %v (err %.2f%%)", got, want, e*100)
	}
}

// 同一条路径用同一把刀再走一遍，第二步去除量必须为 0。
func TestRepeatPathRemovesNothing(t *testing.T) {
	step := Step{
		ToolID: "T1",
		Start:  Vec3{X: 10, Y: 15, Z: 30},
		Moves: []Move{
			{Type: MoveRapid, To: Vec3{X: 10, Y: 15, Z: 16}},
			{Type: MoveLinear, To: Vec3{X: 50, Y: 15, Z: 16}},
		},
	}
	job := Job{
		Stock: Stock{Type: StockBox, Min: Vec3{0, 0, 0}, Max: Vec3{60, 30, 20}},
		Tools: []Tool{{ID: "T1", Type: ToolFlat, Diameter: 10, FluteLength: 20}},
		Steps: []Step{step, step},
	}
	res, err := Run(job)
	if err != nil {
		t.Fatal(err)
	}
	if res.Steps[0].RemovedVolume <= 0 {
		t.Fatalf("first pass should remove material, got %v", res.Steps[0].RemovedVolume)
	}
	if got := res.Steps[1].RemovedVolume; got != 0 {
		t.Fatalf("second pass removed %v, want 0", got)
	}
}

// 球头刀直槽：截面为矩形加深处半球形槽底，去除量应小于同刀尖深度的平底刀。
func TestBallNoseSlotLessThanFlat(t *testing.T) {
	mk := func(tt ToolType) Job {
		return Job{
			Stock: Stock{Type: StockBox, Min: Vec3{0, 0, 0}, Max: Vec3{60, 30, 20}},
			Tools: []Tool{{ID: "T1", Type: tt, Diameter: 10, FluteLength: 20}},
			Steps: []Step{{
				ToolID: "T1",
				Start:  Vec3{X: 10, Y: 15, Z: 30},
				Moves: []Move{
					{Type: MoveRapid, To: Vec3{X: 10, Y: 15, Z: 16}},
					{Type: MoveLinear, To: Vec3{X: 50, Y: 15, Z: 16}},
				},
			}},
		}
	}
	flat, err := Run(mk(ToolFlat))
	if err != nil {
		t.Fatal(err)
	}
	ball, err := Run(mk(ToolBall))
	if err != nil {
		t.Fatal(err)
	}
	f, b := flat.Steps[0].RemovedVolume, ball.Steps[0].RemovedVolume
	if b <= 0 || b >= f {
		t.Fatalf("ball removed %v, want 0 < ball < flat(%v)", b, f)
	}
	// 解析值：中段棱柱 L×球缺截面积 + 一个完整球缺（两端各半个）。
	// 槽深 h=4 < R=5，球缺截面积 = R²·acos((R-h)/R) − (R−h)·√(2Rh−h²)。
	const R, h, L = 5.0, 4.0, 40.0
	segArea := R*R*math.Acos((R-h)/R) - (R-h)*math.Sqrt(2*R*h-h*h)
	capVol := math.Pi * h * h * (R - h/3)
	want := L*segArea + capVol
	if e := math.Abs(b-want) / want; e > 0.03 {
		t.Fatalf("ball removed %v, want %v (err %.2f%%)", b, want, e*100)
	}
}
