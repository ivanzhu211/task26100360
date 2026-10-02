package millsim

import "fmt"

// DefaultResolution 是默认的 dexel 网格边长（mm）。
const DefaultResolution = 0.2

// Run 执行整份程序的仿真，返回每步去除量与剩余量。
func Run(job Job) (Result, error) {
	if err := validate(job); err != nil {
		return Result{}, err
	}
	cell := job.Resolution
	if cell <= 0 {
		cell = DefaultResolution
	}
	tools := make(map[string]Tool, len(job.Tools))
	for _, t := range job.Tools {
		tools[t.ID] = t
	}
	field := NewField(job.Stock, cell)
	res := Result{InitialVolume: field.Volume()}
	remaining := res.InitialVolume
	for i, step := range job.Steps {
		tool, ok := tools[step.ToolID]
		if !ok {
			return Result{}, fmt.Errorf("step %d: unknown tool %q", i, step.ToolID)
		}
		pos := step.Start
		for _, m := range step.Moves {
			pos = sweepMove(field, tool, pos, m)
		}
		remaining = field.Volume()
		res.Steps = append(res.Steps, StepResult{
			Index:           i,
			RemovedVolume:   res.InitialVolume - remaining - removedSoFar(res.Steps),
			RemainingVolume: remaining,
		})
	}
	res.FinalVolume = remaining
	return res, nil
}

func removedSoFar(steps []StepResult) float64 {
	var v float64
	for _, s := range steps {
		v += s.RemovedVolume
	}
	return v
}

func validate(job Job) error {
	s := job.Stock
	switch s.Type {
	case StockBox:
		if s.Max.X <= s.Min.X || s.Max.Y <= s.Min.Y || s.Max.Z <= s.Min.Z {
			return fmt.Errorf("invalid box stock: max must be greater than min")
		}
	case StockCylinder:
		if s.Radius <= 0 || s.ZMax <= s.ZMin {
			return fmt.Errorf("invalid cylinder stock")
		}
	default:
		return fmt.Errorf("unknown stock type %q", s.Type)
	}
	for _, t := range job.Tools {
		if t.Type != ToolFlat && t.Type != ToolBall {
			return fmt.Errorf("tool %q: unknown type %q", t.ID, t.Type)
		}
		if t.Diameter <= 0 || t.FluteLength <= 0 {
			return fmt.Errorf("tool %q: diameter and fluteLength must be positive", t.ID)
		}
	}
	for i, step := range job.Steps {
		for j, m := range step.Moves {
			switch m.Type {
			case MoveRapid, MoveLinear:
			case MoveArc:
				if m.Direction != ArcCW && m.Direction != ArcCCW {
					return fmt.Errorf("step %d move %d: arc direction must be cw or ccw", i, j)
				}
			default:
				return fmt.Errorf("step %d move %d: unknown move type %q", i, j, m.Type)
			}
		}
	}
	return nil
}
