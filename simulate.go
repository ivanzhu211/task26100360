package millsim

import "fmt"

// defaultResolution is the grid cell size used when the job does not
// specify one, in millimeters.
const defaultResolution = 0.5

// Simulate runs the whole job and returns the per-step result.
func Simulate(job Job) (Result, error) {
	if err := validate(job); err != nil {
		return Result{}, err
	}
	res := job.Resolution
	if res <= 0 {
		res = defaultResolution
	}
	h := newHeightfield(job.Stock, res)

	tools := make(map[string]tool, len(job.Tools))
	for name, def := range job.Tools {
		tools[name] = newTool(def)
	}

	out := Result{InitialVolume: h.volume()}
	for _, step := range job.Steps {
		t, ok := tools[step.Tool]
		if !ok {
			return Result{}, fmt.Errorf("step %q: unknown tool %q", step.Name, step.Tool)
		}
		pos := step.Start
		removed := 0.0
		for _, m := range step.Moves {
			switch m.Type {
			case "rapid", "linear":
				removed += h.sweepLine(t, pos, m.To)
			case "arc":
				removed += h.sweepArc(t, pos, m)
			default:
				return Result{}, fmt.Errorf("step %q: unknown move type %q", step.Name, m.Type)
			}
			pos = m.To
		}
		out.Steps = append(out.Steps, StepResult{
			Name:            step.Name,
			RemovedVolume:   removed,
			RemainingVolume: h.volume(),
		})
	}
	out.FinalVolume = h.volume()
	return out, nil
}

func validate(job Job) error {
	switch job.Stock.Type {
	case "box":
		if job.Stock.Size[0] <= 0 || job.Stock.Size[1] <= 0 || job.Stock.Size[2] <= 0 {
			return fmt.Errorf("box stock: size must be positive, got %v", job.Stock.Size)
		}
	case "cylinder":
		if job.Stock.Radius <= 0 || job.Stock.Height <= 0 {
			return fmt.Errorf("cylinder stock: radius and height must be positive")
		}
	default:
		return fmt.Errorf("unknown stock type %q", job.Stock.Type)
	}
	for name, d := range job.Tools {
		switch d.Type {
		case "flat", "ball":
		default:
			return fmt.Errorf("tool %q: unknown type %q", name, d.Type)
		}
		if d.Diameter <= 0 {
			return fmt.Errorf("tool %q: diameter must be positive", name)
		}
		if d.FluteLength <= 0 {
			return fmt.Errorf("tool %q: fluteLength must be positive", name)
		}
	}
	for _, s := range job.Steps {
		if _, ok := job.Tools[s.Tool]; !ok {
			return fmt.Errorf("step %q: unknown tool %q", s.Name, s.Tool)
		}
	}
	return nil
}
