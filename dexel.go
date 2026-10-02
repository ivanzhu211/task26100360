package millsim

import "math"

// interval 表示一段竖直材料区间 [Lo, Hi]。
type interval struct {
	Lo, Hi float64
}

// Field 是毛坯的 dexel（竖直柱）表示：XY 规则网格上每格保存若干材料区间。
type Field struct {
	xmin, ymin float64
	cell       float64
	nx, ny     int
	cols       [][]interval // 长度 nx*ny
}

// NewField 按给定分辨率从毛坯构建 dexel 场。
func NewField(s Stock, cell float64) *Field {
	var xmin, ymin, xmax, ymax, zmin, zmax float64
	switch s.Type {
	case StockBox:
		xmin, ymin = s.Min.X, s.Min.Y
		xmax, ymax = s.Max.X, s.Max.Y
		zmin, zmax = s.Min.Z, s.Max.Z
	case StockCylinder:
		xmin, ymin = s.Center[0]-s.Radius, s.Center[1]-s.Radius
		xmax, ymax = s.Center[0]+s.Radius, s.Center[1]+s.Radius
		zmin, zmax = s.ZMin, s.ZMax
	}
	nx := int(math.Ceil((xmax - xmin) / cell))
	ny := int(math.Ceil((ymax - ymin) / cell))
	f := &Field{xmin: xmin, ymin: ymin, cell: cell, nx: nx, ny: ny, cols: make([][]interval, nx*ny)}
	for iy := 0; iy < ny; iy++ {
		cy := ymin + (float64(iy)+0.5)*cell
		for ix := 0; ix < nx; ix++ {
			cx := xmin + (float64(ix)+0.5)*cell
			if !insideFootprint(s, cx, cy) {
				continue
			}
			f.cols[iy*nx+ix] = []interval{{Lo: zmin, Hi: zmax}}
		}
	}
	return f
}

func insideFootprint(s Stock, x, y float64) bool {
	switch s.Type {
	case StockBox:
		return x >= s.Min.X && x <= s.Max.X && y >= s.Min.Y && y <= s.Max.Y
	case StockCylinder:
		dx, dy := x-s.Center[0], y-s.Center[1]
		return dx*dx+dy*dy <= s.Radius*s.Radius
	}
	return false
}

// Volume 返回当前材料总体积（mm^3）。
func (f *Field) Volume() float64 {
	var v float64
	for _, col := range f.cols {
		for _, in := range col {
			v += in.Hi - in.Lo
		}
	}
	return v * f.cell * f.cell
}

// subtractTool 在刀尖位于 p 时，从所有被刀具切削部分覆盖的格子中减去材料。
func (f *Field) subtractTool(p Vec3, t Tool) {
	R := t.Diameter / 2
	zTop := p.Z + t.FluteLength // 刃长以上不切料
	ix0 := clampi(int(math.Floor((p.X-R-f.xmin)/f.cell)), 0, f.nx-1)
	ix1 := clampi(int(math.Floor((p.X+R-f.xmin)/f.cell)), 0, f.nx-1)
	iy0 := clampi(int(math.Floor((p.Y-R-f.ymin)/f.cell)), 0, f.ny-1)
	iy1 := clampi(int(math.Floor((p.Y+R-f.ymin)/f.cell)), 0, f.ny-1)
	for iy := iy0; iy <= iy1; iy++ {
		cy := f.ymin + (float64(iy)+0.5)*f.cell
		for ix := ix0; ix <= ix1; ix++ {
			cx := f.xmin + (float64(ix)+0.5)*f.cell
			dx, dy := cx-p.X, cy-p.Y
			rho2 := dx*dx + dy*dy
			if rho2 > R*R {
				continue
			}
			zBot := p.Z
			if t.Type == ToolBall {
				// 球头：距轴线 rho 处切削底面在半球面上
				zBot = p.Z + R - math.Sqrt(R*R-rho2)
			}
			idx := iy*f.nx + ix
			f.cols[idx] = subtractInterval(f.cols[idx], zBot, zTop)
		}
	}
}

// subtractInterval 从区间列表中去掉 [lo, hi]。
func subtractInterval(cols []interval, lo, hi float64) []interval {
	out := cols[:0]
	for _, in := range cols {
		if hi <= in.Lo || lo >= in.Hi {
			out = append(out, in)
			continue
		}
		if in.Lo < lo {
			out = append(out, interval{Lo: in.Lo, Hi: lo})
		}
		if in.Hi > hi {
			out = append(out, interval{Lo: hi, Hi: in.Hi})
		}
	}
	return out
}

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
