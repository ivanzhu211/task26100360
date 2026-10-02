package millsim

import "math"

// sweepMove 把从 from 出发的一个移动扫掠到 dexel 场上，返回结束位置。
func sweepMove(f *Field, t Tool, from Vec3, m Move) Vec3 {
	switch m.Type {
	case MoveArc:
		sweepArc(f, t, from, m)
	default: // rapid 与 linear 都是直线扫掠
		sweepLine(f, t, from, m.To)
	}
	return m.To
}

// sweepLine 沿直线按不超过半步长的间距放置刀具。
func sweepLine(f *Field, t Tool, from, to Vec3) {
	d := dist3(from, to)
	n := stepsFor(d, f.cell)
	for i := 1; i <= n; i++ {
		s := float64(i) / float64(n)
		f.subtractTool(lerp3(from, to, s), t)
	}
}

// sweepArc 把 XY 平面圆弧（可带 Z 线性变化，即螺旋）细分为小线段扫掠。
// 起点与终点在 XY 上重合时按整圆处理。
func sweepArc(f *Field, t Tool, from Vec3, m Move) {
	cx, cy := m.Center[0], m.Center[1]
	r := math.Hypot(from.X-cx, from.Y-cy)
	if r < 1e-9 { // 半径为 0，退化为竖直直线
		sweepLine(f, t, from, m.To)
		return
	}
	a0 := math.Atan2(from.Y-cy, from.X-cx)
	var sweep float64
	if math.Hypot(m.To.X-from.X, m.To.Y-from.Y) < 1e-9 {
		sweep = 2 * math.Pi // 整圆
	} else {
		a1 := math.Atan2(m.To.Y-cy, m.To.X-cx)
		sweep = math.Mod(a1-a0, 2*math.Pi)
		if sweep <= 0 {
			sweep += 2 * math.Pi
		}
	}
	if m.Direction == ArcCW {
		sweep = -sweep
	}
	arcLen := math.Abs(sweep) * r
	zLen := math.Abs(m.To.Z - from.Z)
	n := stepsFor(math.Hypot(arcLen, zLen), f.cell)
	for i := 1; i <= n; i++ {
		s := float64(i) / float64(n)
		a := a0 + sweep*s
		f.subtractTool(Vec3{
			X: cx + r*math.Cos(a),
			Y: cy + r*math.Sin(a),
			Z: from.Z + (m.To.Z-from.Z)*s,
		}, t)
	}
}

// stepsFor 按半步长（半格）计算细分段数。
func stepsFor(d, cell float64) int {
	n := int(math.Ceil(d / (0.5 * cell)))
	if n < 1 {
		n = 1
	}
	return n
}

func dist3(a, b Vec3) float64 {
	return math.Hypot(math.Hypot(a.X-b.X, a.Y-b.Y), a.Z-b.Z)
}

func lerp3(a, b Vec3, s float64) Vec3 {
	return Vec3{X: a.X + (b.X-a.X)*s, Y: a.Y + (b.Y-a.Y)*s, Z: a.Z + (b.Z-a.Z)*s}
}
