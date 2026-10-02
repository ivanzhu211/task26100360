package millsim

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadJob 从 JSON 文件读取仿真输入。
func LoadJob(path string) (Job, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Job{}, fmt.Errorf("read job: %w", err)
	}
	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return Job{}, fmt.Errorf("parse job %s: %w", path, err)
	}
	return job, nil
}

// SaveResult 把仿真结果写入 JSON 文件（带缩进，便于人工查看）。
func SaveResult(path string, r Result) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("encode result: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write result: %w", err)
	}
	return nil
}

// LoadResult 从 JSON 文件读回仿真结果（主要用于校验与测试）。
func LoadResult(path string) (Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Result{}, fmt.Errorf("read result: %w", err)
	}
	var r Result
	if err := json.Unmarshal(data, &r); err != nil {
		return Result{}, fmt.Errorf("parse result %s: %w", path, err)
	}
	return r, nil
}
