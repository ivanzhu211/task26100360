// millsim 命令行工具：从 JSON 文件读入毛坯/刀具/刀路，输出每步去除体积。
//
// 用法：
//
//	millsim -job program.json -out result.json
package main

import (
	"flag"
	"fmt"
	"os"

	"millsim"
)

func main() {
	jobPath := flag.String("job", "", "输入 JSON 文件（毛坯、刀具、刀路）")
	outPath := flag.String("out", "", "输出 JSON 文件（每步去除量与剩余量）")
	flag.Parse()
	if *jobPath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "usage: millsim -job <input.json> -out <result.json>")
		os.Exit(2)
	}
	job, err := millsim.LoadJob(*jobPath)
	if err != nil {
		fatal(err)
	}
	res, err := millsim.Run(job)
	if err != nil {
		fatal(err)
	}
	if err := millsim.SaveResult(*outPath, res); err != nil {
		fatal(err)
	}
	fmt.Printf("initial=%.3f mm^3 final=%.3f mm^3 steps=%d\n",
		res.InitialVolume, res.FinalVolume, len(res.Steps))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "millsim:", err)
	os.Exit(1)
}
