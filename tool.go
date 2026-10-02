package millsim

import "math"

// tool is the runtime form of ToolDef.
type tool struct {
	radius      float64
	fluteLength float64
	ball        bool
}

func newTool(d ToolDef) tool {
	return tool{
		radius:      d.Diameter / 2,
		fluteLength: d.FluteLength,
		ball:        d.Type == "ball",
	}
}

// surfaceZ returns the Z of the tool's bottom surface at horizontal
// distance d from the tool axis, given the tip height tipZ. For a flat
// end mill the bottom face is flat; for a ball nose the bottom is a
// hemisphere of the tool radius.
func (t tool) surfaceZ(tipZ, d float64) float64 {
	if !t.ball {
		return tipZ
	}
	s := t.radius*t.radius - d*d
	if s <= 0 {
		return tipZ + t.radius
	}
	return tipZ + t.radius - math.Sqrt(s)
}
