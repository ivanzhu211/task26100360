// Command millsim runs a 3-axis milling simulation: it reads the
// stock, tools and toolpath from a JSON file and writes the removed
// and remaining volumes per step to another JSON file.
//
// Usage: millsim <input.json> <output.json>
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"millsim"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: millsim <input.json> <output.json>")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fatal(err)
	}
	var job millsim.Job
	if err := json.Unmarshal(data, &job); err != nil {
		fatal(fmt.Errorf("parse %s: %w", os.Args[1], err))
	}
	res, err := millsim.Simulate(job)
	if err != nil {
		fatal(err)
	}
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fatal(err)
	}
	out = append(out, '\n')
	if err := os.WriteFile(os.Args[2], out, 0o644); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "millsim:", err)
	os.Exit(1)
}
