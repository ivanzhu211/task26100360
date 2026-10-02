package millsim

import "math"

// heightfield is a regular XY grid of material columns. Each column
// stores the Z of its top surface; material fills [baseZ, top].
// Columns outside the stock footprint have top == baseZ (empty).
type heightfield struct {
	res    float64
	nx, ny int
	ox, oy float64 // grid min corner
	baseZ  float64
	top    []float64
}

// newHeightfield builds a heightfield covering the stock bounding box.
func newHeightfield(s StockDef, res float64) *heightfield {
	var minX, minY, w, d, topZ, baseZ float64
	switch s.Type {
	case "box":
		minX, minY = s.Origin[0], s.Origin[1]
		w, d = s.Size[0], s.Size[1]
		baseZ = s.Origin[2]
		topZ = baseZ + s.Size[2]
	case "cylinder":
		minX = s.Origin[0] - s.Radius
		minY = s.Origin[1] - s.Radius
		w, d = 2*s.Radius, 2*s.Radius
		baseZ = s.Origin[2]
		topZ = baseZ + s.Height
	}
	nx := int(math.Ceil(w / res))
	ny := int(math.Ceil(d / res))
	h := &heightfield{
		res:   res,
		nx:    nx,
		ny:    ny,
		ox:    minX,
		oy:    minY,
		baseZ: baseZ,
		top:   make([]float64, nx*ny),
	}
	for j := 0; j < ny; j++ {
		for i := 0; i < nx; i++ {
			x, y := h.cellCenter(i, j)
			if insideStock(s, x, y) {
				h.top[j*nx+i] = topZ
			} else {
				h.top[j*nx+i] = baseZ
			}
		}
	}
	return h
}

func insideStock(s StockDef, x, y float64) bool {
	switch s.Type {
	case "box":
		return x >= s.Origin[0] && x <= s.Origin[0]+s.Size[0] &&
			y >= s.Origin[1] && y <= s.Origin[1]+s.Size[1]
	case "cylinder":
		dx := x - s.Origin[0]
		dy := y - s.Origin[1]
		return dx*dx+dy*dy <= s.Radius*s.Radius
	}
	return false
}

func (h *heightfield) cellCenter(i, j int) (float64, float64) {
	return h.ox + (float64(i)+0.5)*h.res, h.oy + (float64(j)+0.5)*h.res
}

// volume returns the current material volume in mm^3.
func (h *heightfield) volume() float64 {
	sum := 0.0
	for _, z := range h.top {
		sum += z - h.baseZ
	}
	return sum * h.res * h.res
}

// cutAt lowers the column (i, j) to z if the tool surface at z cuts
// it: the column top must be above z and within reach of the flutes.
// Material above z+fluteLength cannot be represented in a heightfield
// (it would float), so such columns are left untouched.
// Returns the removed volume for this column.
func (h *heightfield) cutAt(i, j int, z, fluteLength float64) float64 {
	idx := j*h.nx + i
	top := h.top[idx]
	if top <= z || top > z+fluteLength {
		return 0
	}
	h.top[idx] = z
	return (top - z) * h.res * h.res
}
