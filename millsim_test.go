package millsim

import (
	"math"
	"testing"
)

func runJob(t *testing.T, job Job) Result {
	t.Helper()
	res, err := Simulate(job)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}
	return res
}

func assertClose(t *testing.T, name string, got, want, tolFrac float64) {
	t.Helper()
	if want == 0 {
		if got != 0 {
			t.Fatalf("%s: got %v, want 0", name, got)
		}
		return
	}
	diff := math.Abs(got-want) / math.Abs(want)
	if diff > tolFrac {
		t.Fatalf("%s: got %v, want %v (rel err %.4f > %.4f)", name, got, want, diff, tolFrac)
	}
	t.Logf("%s: got %v, want %v (rel err %.4f)", name, got, want, diff)
}

// Flat end mill cutting a straight slot: removed volume equals the
// swept capsule footprint times depth.
func TestFlatEndmillStraightSlot(t *testing.T) {
	job := Job{
		Stock: StockDef{Type: "box", Origin: Vec3{0, 0, 0}, Size: Vec3{100, 50, 20}},
		Tools: map[string]ToolDef{
			"T1": {Type: "flat", Diameter: 10, FluteLength: 30},
		},
		Resolution: 0.25,
		Steps: []Step{{
			Name:  "slot",
			Tool:  "T1",
			Start: Vec3{5, 25, 30},
			Moves: []Move{
				{Type: "rapid", To: Vec3{5, 25, 10}},
				{Type: "linear", To: Vec3{95, 25, 10}},
			},
		}},
	}
	res := runJob(t, job)

	assertClose(t, "initial volume", res.InitialVolume, 100*50*20, 1e-9)
	// Capsule footprint: 90x10 rectangle + two r=5 semicircles, 10 deep.
	want := (90*10 + math.Pi*25) * 10
	assertClose(t, "removed volume", res.Steps[0].RemovedVolume, want, 0.02)
	assertClose(t, "remaining volume", res.Steps[0].RemainingVolume, 100000-want, 0.02)
}

// Flat end mill cutting a full-circle arc slot: removed volume equals
// the annulus area times depth.
func TestFlatEndmillFullCircleSlot(t *testing.T) {
	job := Job{
		Stock: StockDef{Type: "box", Origin: Vec3{-20, -20, 0}, Size: Vec3{40, 40, 20}},
		Tools: map[string]ToolDef{
			"T1": {Type: "flat", Diameter: 10, FluteLength: 30},
		},
		Resolution: 0.2,
		Steps: []Step{{
			Name:  "circle",
			Tool:  "T1",
			Start: Vec3{10, 0, 30},
			Moves: []Move{
				{Type: "rapid", To: Vec3{10, 0, 15}},
				// Start == end in XY: full circle.
				{Type: "arc", Center: [2]float64{0, 0}, CW: false, To: Vec3{10, 0, 15}},
			},
		}},
	}
	res := runJob(t, job)

	// Annulus rho in [5, 15], 5 mm deep.
	want := math.Pi * (15*15 - 5*5) * 5
	assertClose(t, "removed volume", res.Steps[0].RemovedVolume, want, 0.02)
}

// Helical plunge with a flat end mill. The expected volume is
// computed by independent numerical integration of the swept
// envelope: for each ground point (rho, phi) in the annulus, the
// removed depth is the pitch times the largest covered helix angle.
func TestFlatEndmillHelicalPlunge(t *testing.T) {
	const (
		rPath  = 10.0 // helix radius
		rTool  = 5.0
		pitch  = 2.0 // Z drop per revolution
		stockT = 20.0
	)
	job := Job{
		Stock: StockDef{Type: "box", Origin: Vec3{-20, -20, 0}, Size: Vec3{40, 40, stockT}},
		Tools: map[string]ToolDef{
			"T1": {Type: "flat", Diameter: 2 * rTool, FluteLength: 30},
		},
		Resolution: 0.2,
		Steps: []Step{{
			Name:  "helix",
			Tool:  "T1",
			Start: Vec3{rPath, 0, 30},
			Moves: []Move{
				{Type: "rapid", To: Vec3{rPath, 0, stockT}},
				{Type: "arc", Center: [2]float64{0, 0}, CW: false, To: Vec3{rPath, 0, stockT - pitch}},
			},
		}},
	}
	res := runJob(t, job)

	// Independent expectation: integrate removed depth over the
	// annulus in polar coordinates.
	const nR = 4000
	const nP = 4000
	want := 0.0
	dRho := 2 * rTool / nR
	dPhi := 2 * math.Pi / nP
	for ir := 0; ir < nR; ir++ {
		rho := rPath - rTool + (float64(ir)+0.5)*dRho
		cosA := (rPath*rPath + rho*rho - rTool*rTool) / (2 * rPath * rho)
		alpha := math.Acos(math.Min(1, math.Max(-1, cosA)))
		for ip := 0; ip < nP; ip++ {
			phi := (float64(ip) + 0.5) * dPhi
			// The helix is a closed loop in XY (start == end), so a
			// column whose coverage window wraps past theta = 0/2pi
			// is covered at the end of the helix too and gets the
			// full pitch depth.
			var thetaMax float64
			if phi-alpha < 0 || phi+alpha > 2*math.Pi {
				thetaMax = 2 * math.Pi
			} else {
				thetaMax = phi + alpha
			}
			depth := pitch * thetaMax / (2 * math.Pi)
			want += depth * rho * dRho * dPhi
		}
	}
	assertClose(t, "removed volume", res.Steps[0].RemovedVolume, want, 0.03)
}

// Ball nose cutting a straight slot: removed volume equals the
// semicircular groove cross-section times length plus the two
// half-bowl ends.
func TestBallNoseStraightSlot(t *testing.T) {
	const rTool = 5.0
	job := Job{
		Stock: StockDef{Type: "box", Origin: Vec3{0, 0, 0}, Size: Vec3{60, 20, 20}},
		Tools: map[string]ToolDef{
			"T1": {Type: "ball", Diameter: 2 * rTool, FluteLength: 30},
		},
		Resolution: 0.2,
		Steps: []Step{{
			Name:  "slot",
			Tool:  "T1",
			Start: Vec3{10, 10, 30},
			Moves: []Move{
				{Type: "rapid", To: Vec3{10, 10, 15}},
				{Type: "linear", To: Vec3{50, 10, 15}},
			},
		}},
	}
	res := runJob(t, job)

	// Groove cross-section: area under the semicircle = pi r^2 / 2.
	groove := math.Pi * rTool * rTool / 2 * 40
	// Each of the two ends contributes a half "bowl": the integral of
	// sqrt(r^2-d^2) over a half disk = pi r^3 / 3.
	bowls := 2 * math.Pi * rTool * rTool * rTool / 3
	want := groove + bowls
	assertClose(t, "removed volume", res.Steps[0].RemovedVolume, want, 0.03)
}

// Running the same path twice with the same tool removes nothing the
// second time.
func TestRepeatPathRemovesNothing(t *testing.T) {
	step := Step{
		Name:  "pass",
		Tool:  "T1",
		Start: Vec3{5, 25, 30},
		Moves: []Move{
			{Type: "rapid", To: Vec3{5, 25, 10}},
			{Type: "linear", To: Vec3{95, 25, 10}},
		},
	}
	step2 := step
	step2.Name = "pass-again"
	job := Job{
		Stock: StockDef{Type: "box", Origin: Vec3{0, 0, 0}, Size: Vec3{100, 50, 20}},
		Tools: map[string]ToolDef{
			"T1": {Type: "flat", Diameter: 10, FluteLength: 30},
		},
		Resolution: 0.25,
		Steps:      []Step{step, step2},
	}
	res := runJob(t, job)

	if res.Steps[1].RemovedVolume != 0 {
		t.Fatalf("second pass removed %v, want 0", res.Steps[1].RemovedVolume)
	}
	if res.Steps[1].RemainingVolume != res.Steps[0].RemainingVolume {
		t.Fatalf("remaining changed: %v -> %v",
			res.Steps[0].RemainingVolume, res.Steps[1].RemainingVolume)
	}
}

// Cylinder stock: initial volume approximates pi r^2 h.
func TestCylinderStockInitialVolume(t *testing.T) {
	job := Job{
		Stock:      StockDef{Type: "cylinder", Origin: Vec3{0, 0, 0}, Radius: 25, Height: 10},
		Tools:      map[string]ToolDef{"T1": {Type: "flat", Diameter: 10, FluteLength: 30}},
		Resolution: 0.25,
		Steps:      nil,
	}
	res := runJob(t, job)
	assertClose(t, "initial volume", res.InitialVolume, math.Pi*25*25*10, 0.01)
	if res.FinalVolume != res.InitialVolume {
		t.Fatalf("no steps: final %v != initial %v", res.FinalVolume, res.InitialVolume)
	}
}
