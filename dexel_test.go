package millsim

import (
	"math"
	"testing"
)

func TestFieldVolumeBox(t *testing.T) {
	f := NewField(Stock{Type: StockBox, Min: Vec3{0, 0, 0}, Max: Vec3{20, 20, 10}}, 0.2)
	want := 20.0 * 20.0 * 10.0
	if got := f.Volume(); math.Abs(got-want) > 1e-6*want {
		t.Fatalf("box volume = %v, want %v", got, want)
	}
}

func TestFieldVolumeCylinder(t *testing.T) {
	f := NewField(Stock{Type: StockCylinder, Center: [2]float64{0, 0}, Radius: 10, ZMin: 0, ZMax: 10}, 0.1)
	want := math.Pi * 10 * 10 * 10
	got := f.Volume()
	if err := math.Abs(got-want) / want; err > 0.01 {
		t.Fatalf("cylinder volume = %v, want %v (err %.2f%%)", got, want, err*100)
	}
}

func TestSubtractToolRemovesCylinder(t *testing.T) {
	f := NewField(Stock{Type: StockBox, Min: Vec3{0, 0, 0}, Max: Vec3{30, 30, 10}}, 0.1)
	tool := Tool{ID: "T1", Type: ToolFlat, Diameter: 10, FluteLength: 20}
	before := f.Volume()
	f.subtractTool(Vec3{X: 15, Y: 15, Z: 5}, tool)
	removed := before - f.Volume()
	want := math.Pi * 25 * 5 // 半径5、深5的圆柱
	if err := math.Abs(removed-want) / want; err > 0.02 {
		t.Fatalf("removed = %v, want %v (err %.2f%%)", removed, want, err*100)
	}
}

func TestSubtractToolTwiceIsIdempotent(t *testing.T) {
	f := NewField(Stock{Type: StockBox, Min: Vec3{0, 0, 0}, Max: Vec3{30, 30, 10}}, 0.1)
	tool := Tool{ID: "T1", Type: ToolFlat, Diameter: 10, FluteLength: 20}
	f.subtractTool(Vec3{X: 15, Y: 15, Z: 5}, tool)
	after := f.Volume()
	f.subtractTool(Vec3{X: 15, Y: 15, Z: 5}, tool)
	if got := f.Volume(); got != after {
		t.Fatalf("second subtract changed volume: %v -> %v", after, got)
	}
}

func TestFluteLengthLimitsCut(t *testing.T) {
	// 刃长只有 3mm：刀尖在 z=0 时只能切到 z=3，上方材料保留。
	f := NewField(Stock{Type: StockBox, Min: Vec3{0, 0, 0}, Max: Vec3{30, 30, 10}}, 0.1)
	tool := Tool{ID: "T1", Type: ToolFlat, Diameter: 10, FluteLength: 3}
	before := f.Volume()
	f.subtractTool(Vec3{X: 15, Y: 15, Z: 0}, tool)
	removed := before - f.Volume()
	want := math.Pi * 25 * 3
	if err := math.Abs(removed-want) / want; err > 0.02 {
		t.Fatalf("removed = %v, want %v (err %.2f%%)", removed, want, err*100)
	}
}
