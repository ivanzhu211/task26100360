package millsim

// Vec3 is a 3D point/vector in millimeters. Z is up.
type Vec3 [3]float64

// StockDef describes the raw stock before machining.
// Type is "box" (axis-aligned) or "cylinder" (vertical axis).
type StockDef struct {
	Type   string  `json:"type"`
	Origin Vec3    `json:"origin"` // box: min corner; cylinder: axis point at base
	Size   Vec3    `json:"size"`   // box only
	Radius float64 `json:"radius"` // cylinder only
	Height float64 `json:"height"` // cylinder only
}

// ToolDef describes a cutting tool.
// Type is "flat" (flat end mill) or "ball" (ball nose).
// The tool tip is the lowest point of the tool.
type ToolDef struct {
	Type        string  `json:"type"`
	Diameter    float64 `json:"diameter"`
	FluteLength float64 `json:"fluteLength"`
}

// Move is a single motion. Type is "rapid", "linear" or "arc".
// Arcs lie in the XY plane (with optional helical Z change) and are
// defined by Center, direction CW and end point To. A full circle is
// expressed by To == current position (in XY).
type Move struct {
	Type   string     `json:"type"`
	To     Vec3       `json:"to"`
	Center [2]float64 `json:"center,omitempty"`
	CW     bool       `json:"cw,omitempty"`
}

// Step is one machining step: a sequence of moves with a single tool,
// starting from Start (guaranteed above the stock top).
type Step struct {
	Name  string `json:"name"`
	Tool  string `json:"tool"`
	Start Vec3   `json:"start"`
	Moves []Move `json:"moves"`
}

// Job is the full simulation input read from a JSON file.
type Job struct {
	Stock      StockDef           `json:"stock"`
	Tools      map[string]ToolDef `json:"tools"`
	Resolution float64            `json:"resolution"` // grid cell size in mm; 0 => default
	Steps      []Step             `json:"steps"`
}

// StepResult reports the outcome of one step.
type StepResult struct {
	Name            string  `json:"name"`
	RemovedVolume   float64 `json:"removedVolume"`
	RemainingVolume float64 `json:"remainingVolume"`
}

// Result is the simulation output written to a JSON file.
type Result struct {
	InitialVolume float64      `json:"initialVolume"`
	Steps         []StepResult `json:"steps"`
	FinalVolume   float64      `json:"finalVolume"`
}
