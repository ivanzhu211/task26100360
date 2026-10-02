package millsim

import "math"

// sweepLine removes the material swept by the tool moving straight
// from p0 to p1 and returns the removed volume. Rapid and cutting
// moves behave identically: anything the flutes sweep is removed.
func (h *heightfield) sweepLine(t tool, p0, p1 Vec3) float64 {
	r := t.radius
	minX := math.Min(p0[0], p1[0]) - r
	maxX := math.Max(p0[0], p1[0]) + r
	minY := math.Min(p0[1], p1[1]) - r
	maxY := math.Max(p0[1], p1[1]) + r
	i0, i1 := h.xRange(minX, maxX)
	j0, j1 := h.yRange(minY, maxY)

	dx := p1[0] - p0[0]
	dy := p1[1] - p0[1]
	dz := p1[2] - p0[2]
	len2 := dx*dx + dy*dy

	removed := 0.0
	for j := j0; j <= j1; j++ {
		for i := i0; i <= i1; i++ {
			cx, cy := h.cellCenter(i, j)
			// Interval of segment parameters t where the column
			// center is within the tool radius of the axis.
			ta, tb, ok := coveredInterval(cx-p0[0], cy-p0[1], dx, dy, len2, r)
			if !ok {
				continue
			}
			var zsurf float64
			if !t.ball {
				// Flat end mill: bottom face is flat, so the
				// lowest surface over the covered interval is at
				// the end with the lower tip Z.
				te := ta
				if dz < 0 {
					te = tb
				}
				zsurf = p0[2] + te*dz
			} else {
				// Ball nose: minimize tipZ(t)+profile(d(t)) over
				// a few candidate parameters.
				t0 := closestParam(cx-p0[0], cy-p0[1], dx, dy, len2)
				zsurf = math.Inf(1)
				for _, te := range [3]float64{ta, t0, tb} {
					d := distAt(cx-p0[0], cy-p0[1], dx, dy, te)
					z := t.surfaceZ(p0[2]+te*dz, d)
					if z < zsurf {
						zsurf = z
					}
				}
			}
			removed += h.cutAt(i, j, zsurf, t.fluteLength)
		}
	}
	return removed
}

// closestParam returns the segment parameter of the point closest to
// (px, py), where the segment starts at the origin with direction
// (dx, dy).
func closestParam(px, py, dx, dy, len2 float64) float64 {
	if len2 == 0 {
		return 0
	}
	t := (px*dx + py*dy) / len2
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

func distAt(px, py, dx, dy, t float64) float64 {
	ex := px - t*dx
	ey := py - t*dy
	return math.Hypot(ex, ey)
}

// coveredInterval returns the parameter interval [ta, tb] of the
// segment during which the point (px, py) is within radius r of the
// segment axis.
func coveredInterval(px, py, dx, dy, len2, r float64) (float64, float64, bool) {
	if len2 == 0 {
		if px*px+py*py <= r*r {
			return 0, 1, true
		}
		return 0, 0, false
	}
	// |p - t*d|^2 = r^2  ->  a t^2 + b t + c = 0
	a := len2
	b := -2 * (px*dx + py*dy)
	c := px*px + py*py - r*r
	disc := b*b - 4*a*c
	if disc < 0 {
		return 0, 0, false
	}
	sq := math.Sqrt(disc)
	ta := (-b - sq) / (2 * a)
	tb := (-b + sq) / (2 * a)
	if tb < 0 || ta > 1 {
		return 0, 0, false
	}
	if ta < 0 {
		ta = 0
	}
	if tb > 1 {
		tb = 1
	}
	return ta, tb, true
}

// sweepArc removes the material swept along a circular arc in the XY
// plane (with optional helical Z change) from p0 to m.To, around
// m.Center. CW selects the clockwise direction seen from +Z. If the
// start and end coincide in XY the arc is a full circle.
func (h *heightfield) sweepArc(t tool, p0 Vec3, m Move) float64 {
	cx, cy := m.Center[0], m.Center[1]
	r := math.Hypot(p0[0]-cx, p0[1]-cy)
	a0 := math.Atan2(p0[1]-cy, p0[0]-cx)
	a1 := math.Atan2(m.To[1]-cy, m.To[0]-cx)

	full := math.Hypot(m.To[0]-p0[0], m.To[1]-p0[1]) < 1e-9
	var da float64
	switch {
	case full && m.CW:
		da = -2 * math.Pi
	case full:
		da = 2 * math.Pi
	case m.CW:
		da = a1 - a0
		for da >= 0 {
			da -= 2 * math.Pi
		}
	default:
		da = a1 - a0
		for da <= 0 {
			da += 2 * math.Pi
		}
	}

	// Subdivide so the chord sagitta stays below a fraction of the
	// grid resolution.
	tol := h.res * 0.25
	daMax := math.Pi / 4
	if r > tol {
		daMax = 2 * math.Acos(1-tol/r)
		if daMax > math.Pi/4 {
			daMax = math.Pi / 4
		}
	}
	n := int(math.Ceil(math.Abs(da) / daMax))
	if n < 1 {
		n = 1
	}

	removed := 0.0
	prev := p0
	for k := 1; k <= n; k++ {
		f := float64(k) / float64(n)
		a := a0 + da*f
		next := Vec3{
			cx + r*math.Cos(a),
			cy + r*math.Sin(a),
			p0[2] + (m.To[2]-p0[2])*f,
		}
		removed += h.sweepLine(t, prev, next)
		prev = next
	}
	return removed
}

func (h *heightfield) xRange(minX, maxX float64) (int, int) {
	i0 := int(math.Floor((minX - h.ox) / h.res))
	i1 := int(math.Floor((maxX - h.ox) / h.res))
	if i0 < 0 {
		i0 = 0
	}
	if i1 > h.nx-1 {
		i1 = h.nx - 1
	}
	return i0, i1
}

func (h *heightfield) yRange(minY, maxY float64) (int, int) {
	j0 := int(math.Floor((minY - h.oy) / h.res))
	j1 := int(math.Floor((maxY - h.oy) / h.res))
	if j0 < 0 {
		j0 = 0
	}
	if j1 > h.ny-1 {
		j1 = h.ny - 1
	}
	return j0, j1
}
