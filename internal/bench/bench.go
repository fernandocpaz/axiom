// Package bench contains the model-facing pass@1 benchmark runner. It keeps
// generation separate from compilation so tests can use a deterministic fake
// model while production runs use an explicit host adapter.
package bench

import (
	"encoding/json"
	"fmt"

	"github.com/fernandocpaz/axiom/internal/checker"
	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
)

type Task struct {
	ID     string `json:"id"`
	Prompt string `json:"prompt"`
}

type Result struct {
	ID        string `json:"id"`
	Samples   int    `json:"samples"`
	PassAt1   bool   `json:"pass_at_1"`
	ParseOK   int    `json:"parse_ok"`
	CheckOK   int    `json:"check_ok"`
	LastError string `json:"last_error,omitempty"`
}

type Summary struct {
	Schema  string   `json:"schema"`
	Model   string   `json:"model,omitempty"`
	Tasks   int      `json:"tasks"`
	PassAt1 int      `json:"pass_at_1"`
	Results []Result `json:"results"`
}

const Schema = "axiom.bench/v1"

func DecodeManifest(data []byte) ([]Task, error) {
	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("decode benchmark manifest: %w", err)
	}
	for i, task := range tasks {
		if task.ID == "" {
			task.ID = fmt.Sprintf("task-%d", i+1)
			tasks[i] = task
		}
		if task.Prompt == "" {
			return nil, fmt.Errorf("benchmark task %q has empty prompt", task.ID)
		}
	}
	return tasks, nil
}

func Run(tasks []Task, samples int, model string, generate func(string) (string, error)) Summary {
	if samples < 1 {
		samples = 1
	}
	summary := Summary{Schema: Schema, Model: model, Tasks: len(tasks), Results: make([]Result, 0, len(tasks))}
	for _, task := range tasks {
		result := Result{ID: task.ID, Samples: samples}
		for i := 0; i < samples; i++ {
			source, err := generate(task.Prompt)
			if err != nil {
				result.LastError = err.Error()
				continue
			}
			tokens, err := lexer.Scan(source)
			if err != nil {
				result.LastError = err.Error()
				continue
			}
			result.ParseOK++
			program, err := parser.Parse(tokens)
			if err != nil {
				result.LastError = err.Error()
				continue
			}
			if diagnostics := checker.Check(program); len(diagnostics) != 0 {
				result.LastError = diagnostics[0].Error()
				continue
			}
			result.CheckOK++
		}
		result.PassAt1 = result.CheckOK > 0
		if result.PassAt1 {
			summary.PassAt1++
		}
		summary.Results = append(summary.Results, result)
	}
	return summary
}
