package millsim

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJobAndSaveResult(t *testing.T) {
	dir := t.TempDir()
	jobPath := filepath.Join(dir, "job.json")
	outPath := filepath.Join(dir, "result.json")
	data := []byte(`{
	  "stock": {"type": "box", "min": {"x":0,"y":0,"z":0}, "max": {"x":40,"y":30,"z":15}},
	  "tools": [{"id":"T1","type":"flat","diameter":10,"fluteLength":20}],
	  "steps": [{
	    "toolId": "T1",
	    "start": {"x":10,"y":15,"z":25},
	    "moves": [
	      {"type":"rapid","to":{"x":10,"y":15,"z":12}},
	      {"type":"linear","to":{"x":30,"y":15,"z":12}}
	    ]
	  }],
	  "resolution": 0.2
	}`)
	if err := os.WriteFile(jobPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	job, err := LoadJob(jobPath)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Run(job)
	if err != nil {
		t.Fatal(err)
	}
	if res.InitialVolume != 40*30*15 {
		t.Fatalf("initial volume = %v, want %v", res.InitialVolume, 40*30*15)
	}
	if len(res.Steps) != 1 || res.Steps[0].RemovedVolume <= 0 {
		t.Fatalf("unexpected steps: %+v", res.Steps)
	}
	if err := SaveResult(outPath, res); err != nil {
		t.Fatal(err)
	}
	back, err := LoadResult(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if back.FinalVolume != res.FinalVolume || len(back.Steps) != 1 {
		t.Fatalf("round trip mismatch: %+v", back)
	}
}

func TestLoadJobBadJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(p, []byte(`{"stock":`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadJob(p); err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}
