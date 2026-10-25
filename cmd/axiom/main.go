package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fernandocpaz/axiom/internal/bench"
	"github.com/fernandocpaz/axiom/internal/canon"
	"github.com/fernandocpaz/axiom/internal/checker"
	"github.com/fernandocpaz/axiom/internal/diagnostic"
	"github.com/fernandocpaz/axiom/internal/host"
	"github.com/fernandocpaz/axiom/internal/ir"
	"github.com/fernandocpaz/axiom/internal/lexer"
	"github.com/fernandocpaz/axiom/internal/parser"
	"github.com/fernandocpaz/axiom/internal/repair"
	"github.com/fernandocpaz/axiom/internal/schema"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "fmt":
		formatFiles(os.Args[2:])
	case "check":
		checkFiles(os.Args[2:])
	case "bench":
		benchmarkFiles(os.Args[2:])
	case "ast":
		astFiles(os.Args[2:])
	case "schema":
		schemaFiles(os.Args[2:])
	case "repair":
		repairFiles(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		if len(os.Args) != 2 {
			usage()
		}
		runFile(os.Args[1])
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: axiom <file.ax>")
	fmt.Fprintln(os.Stderr, "       axiom fmt [-w|-check] <file.ax>...")
	fmt.Fprintln(os.Stderr, "       axiom check [--json] <file.ax>...")
	fmt.Fprintln(os.Stderr, "       axiom bench [file.ax]...")
	fmt.Fprintln(os.Stderr, "       axiom bench --model-url URL --model NAME [--samples N] tasks.json")
	fmt.Fprintln(os.Stderr, "       axiom ast <file.ax>...")
	fmt.Fprintln(os.Stderr, "       axiom schema <file.ax>...")
	fmt.Fprintln(os.Stderr, "       axiom repair [--json] <file.ax>...")
	os.Exit(2)
}

func runFile(path string) {
	p, e := loadProgram(path, map[string]bool{}, map[string]parser.Program{})
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	diagnostics := checker.Check(p)
	for _, d := range diagnostics {
		fmt.Fprintln(os.Stderr, d.Error())
	}
	if len(diagnostics) > 0 {
		os.Exit(1)
	}
	if err := p.RunWith(parser.RunOptions{Out: func(v parser.Value) { fmt.Println(v) }, Host: configuredHost()}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// configuredHost is deliberately environment-driven so source programs remain
// portable and do not contain credentials. Set AXIOM_MODEL_URL, AXIOM_MODEL,
// and optionally AXIOM_API_KEY to enable infer at runtime. Ollama works with
// AXIOM_MODEL_URL=http://127.0.0.1:11434/v1.
func configuredHost() parser.Host {
	baseURL := os.Getenv("AXIOM_MODEL_URL")
	model := os.Getenv("AXIOM_MODEL")
	if baseURL == "" || model == "" {
		return parser.Host{}
	}
	adapter := host.OpenAICompatible{BaseURL: baseURL, APIKey: os.Getenv("AXIOM_API_KEY"), Model: model}
	return parser.Host{Infer: adapter.Infer}
}

// formatFiles emits canonical source. Axiom has no human-formatted state: the
// formatter is the only writer of .ax files, and -check rejects drift.
func formatFiles(args []string) {
	write := false
	verify := false
	var paths []string
	for _, arg := range args {
		switch arg {
		case "-w", "--write":
			write = true
		case "-check", "--check":
			verify = true
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "axiom fmt: unknown flag %q\n", arg)
				os.Exit(2)
			}
			paths = append(paths, arg)
		}
	}
	if len(paths) == 0 || write && verify {
		usage()
	}
	drifted := false
	for _, path := range paths {
		formatted, already, err := canonicalFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		switch {
		case verify:
			if !already {
				fmt.Fprintln(os.Stderr, path)
				drifted = true
			}
		case write:
			if !already {
				mode := os.FileMode(0600)
				if info, err := os.Stat(path); err == nil {
					mode = info.Mode()
				}
				if err := ioutil.WriteFile(path, []byte(formatted), mode); err != nil {
					fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
					os.Exit(1)
				}
			}
		default:
			fmt.Print(formatted)
		}
	}
	if drifted {
		os.Exit(1)
	}
}

// checkFiles runs static diagnostics and rejects any source that is not in
// canonical form.
func checkFiles(paths []string) {
	jsonOutput := false
	var files []string
	for _, path := range paths {
		if path == "--json" {
			jsonOutput = true
			continue
		}
		files = append(files, path)
	}
	if len(files) == 0 {
		usage()
	}
	failed := false
	results := make([]checkResult, 0, len(files))
	for _, path := range files {
		result := checkResult{File: path, OK: true}
		src, err := ioutil.ReadFile(path)
		if err != nil {
			result.OK = false
			result.Error = err.Error()
			if !jsonOutput {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			}
			failed = true
			results = append(results, result)
			continue
		}
		result.Bytes = len(src)
		tokens, err := lexer.Scan(string(src))
		if err != nil {
			result.OK = false
			result.Diagnostics = []diagnostic.Diagnostic{{Code: "lex-error", Message: err.Error()}}
			result.Error = err.Error()
			if !jsonOutput {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			}
			failed = true
			results = append(results, result)
			continue
		}
		result.LexerTokens = tokenCount(tokens)
		program, err := parser.Parse(tokens)
		if err != nil {
			result.OK = false
			result.Diagnostics = []diagnostic.Diagnostic{{Code: "parse-error", Message: err.Error()}}
			result.Error = err.Error()
			if !jsonOutput {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			}
			failed = true
			results = append(results, result)
			continue
		}
		result.Diagnostics = checker.Check(program)
		for _, d := range result.Diagnostics {
			if !jsonOutput {
				fmt.Fprintf(os.Stderr, "%s: %s\n", path, d.Error())
			}
			failed = true
			result.OK = false
		}
		formatted, already, err := canon.Canonical(string(src))
		if err != nil {
			result.OK = false
			result.Error = err.Error()
			if !jsonOutput {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			}
			failed = true
		} else if !already {
			result.OK = false
			result.CanonicalBytes = len(formatted)
			result.CanonicalLexerTokens = tokenCountString(formatted)
			if !jsonOutput {
				fmt.Fprintf(os.Stderr, "%s: source is not canonical; run axiom fmt -w %s\n", path, path)
			}
			failed = true
		} else {
			result.CanonicalBytes = len(formatted)
			result.CanonicalLexerTokens = tokenCountString(formatted)
		}
		results = append(results, result)
	}
	if jsonOutput {
		if err := writeJSON(results); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if failed {
		os.Exit(1)
	}
}

// checkResult is intentionally stable JSON. Models should be able to consume
// it without scraping human-oriented stderr text.
type checkResult struct {
	File                 string                  `json:"file"`
	OK                   bool                    `json:"ok"`
	Bytes                int                     `json:"bytes,omitempty"`
	LexerTokens          int                     `json:"lexer_tokens,omitempty"`
	CanonicalBytes       int                     `json:"canonical_bytes,omitempty"`
	CanonicalLexerTokens int                     `json:"canonical_lexer_tokens,omitempty"`
	Diagnostics          []diagnostic.Diagnostic `json:"diagnostics,omitempty"`
	Error                string                  `json:"error,omitempty"`
}

type benchmarkResult struct {
	File                 string `json:"file"`
	ParseOK              bool   `json:"parse_ok"`
	CheckOK              bool   `json:"check_ok"`
	Canonical            bool   `json:"canonical"`
	Bytes                int    `json:"bytes,omitempty"`
	CanonicalBytes       int    `json:"canonical_bytes,omitempty"`
	LexerTokens          int    `json:"lexer_tokens,omitempty"`
	CanonicalLexerTokens int    `json:"canonical_lexer_tokens,omitempty"`
	Error                string `json:"error,omitempty"`
}

// benchmarkFiles measures deterministic source properties only. Lexer token
// counts are a portable proxy, not BPE counts; model-specific tokenizers must
// be measured separately because token vocabularies differ by model.
func benchmarkFiles(paths []string) {
	modelURL, model, apiKey := os.Getenv("AXIOM_BENCH_MODEL_URL"), os.Getenv("AXIOM_BENCH_MODEL"), os.Getenv("AXIOM_API_KEY")
	samples := 1
	var sourcePaths []string
	for i := 0; i < len(paths); i++ {
		switch paths[i] {
		case "--model-url":
			if i+1 >= len(paths) {
				usage()
			}
			i++
			modelURL = paths[i]
		case "--model":
			if i+1 >= len(paths) {
				usage()
			}
			i++
			model = paths[i]
		case "--api-key":
			if i+1 >= len(paths) {
				usage()
			}
			i++
			apiKey = paths[i]
		case "--samples":
			if i+1 >= len(paths) {
				usage()
			}
			i++
			if _, err := fmt.Sscan(paths[i], &samples); err != nil || samples < 1 {
				fmt.Fprintln(os.Stderr, "axiom bench: --samples must be a positive integer")
				os.Exit(2)
			}
		default:
			sourcePaths = append(sourcePaths, paths[i])
		}
	}
	if modelURL != "" || model != "" {
		modelBenchmark(sourcePaths, modelURL, model, apiKey, samples)
		return
	}
	paths = sourcePaths
	if len(paths) == 0 {
		var err error
		paths, err = filepath.Glob("examples/*.ax")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	sort.Strings(paths)
	results := make([]benchmarkResult, 0, len(paths))
	for _, path := range paths {
		result := benchmarkResult{File: path}
		src, err := ioutil.ReadFile(path)
		if err != nil {
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		result.Bytes = len(src)
		tokens, err := lexer.Scan(string(src))
		if err != nil {
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		result.LexerTokens = tokenCount(tokens)
		program, err := parser.Parse(tokens)
		if err != nil {
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		result.ParseOK = true
		result.CheckOK = len(checker.Check(program)) == 0
		formatted, canonical, err := canon.Canonical(string(src))
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Canonical = canonical
			result.CanonicalBytes = len(formatted)
			result.CanonicalLexerTokens = tokenCountString(formatted)
		}
		results = append(results, result)
	}
	if err := writeJSON(results); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func modelBenchmark(paths []string, modelURL, model, apiKey string, samples int) {
	if len(paths) != 1 {
		fmt.Fprintln(os.Stderr, "axiom bench: model mode requires one JSON task manifest")
		os.Exit(2)
	}
	if modelURL == "" || model == "" {
		fmt.Fprintln(os.Stderr, "axiom bench: model mode requires --model-url and --model")
		os.Exit(2)
	}
	data, err := ioutil.ReadFile(paths[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	tasks, err := bench.DecodeManifest(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	adapter := host.OpenAICompatible{BaseURL: modelURL, APIKey: apiKey, Model: model}
	summary := bench.Run(tasks, samples, model, adapter.Generate)
	if err := writeJSON(summary); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if summary.PassAt1 != summary.Tasks {
		os.Exit(1)
	}
}

// astFiles emits the stable machine-facing AST. It intentionally emits one
// JSON document per input file so a model can process streams of files without
// scraping diagnostics or compiler-internal Go values.
func astFiles(paths []string) {
	if len(paths) == 0 {
		usage()
	}
	for _, path := range paths {
		src, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		tokens, err := lexer.Scan(string(src))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		program, err := parser.Parse(tokens)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		encoded, err := ir.Encode(program)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		if _, err := os.Stdout.Write(append(encoded, '\n')); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func schemaFiles(paths []string) {
	if len(paths) == 0 {
		usage()
	}
	for _, path := range paths {
		program, err := loadProgram(path, map[string]bool{}, map[string]parser.Program{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		encoded, err := schema.Encode(program)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		if _, err := os.Stdout.Write(append(encoded, '\n')); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func repairFiles(args []string) {
	jsonOutput := false
	var paths []string
	for _, arg := range args {
		if arg == "--json" {
			jsonOutput = true
			continue
		}
		paths = append(paths, arg)
	}
	if len(paths) == 0 {
		usage()
	}
	// Repair output is always JSON because it is intended for an automated
	// generation loop; --json is accepted for symmetry with check.
	_ = jsonOutput
	failed := false
	for _, path := range paths {
		source, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			failed = true
			continue
		}
		report := repair.Analyze(string(source))
		if err := writeJSON(map[string]interface{}{"file": path, "report": report}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !report.Valid {
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}

func tokenCount(tokens []lexer.Token) int {
	if len(tokens) == 0 {
		return 0
	}
	if tokens[len(tokens)-1].Kind == lexer.EOF {
		return len(tokens) - 1
	}
	return len(tokens)
}

func tokenCountString(src string) int {
	tokens, err := lexer.Scan(src)
	if err != nil {
		return 0
	}
	return tokenCount(tokens)
}

func writeJSON(value interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

func canonicalFile(path string) (string, bool, error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return "", false, err
	}
	return canon.Canonical(string(src))
}

func loadProgram(path string, active map[string]bool, cache map[string]parser.Program) (parser.Program, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return parser.Program{}, err
	}
	if active[absolute] {
		return parser.Program{}, fmt.Errorf("import cycle involving %q", path)
	}
	if program, ok := cache[absolute]; ok {
		return program, nil
	}
	src, err := ioutil.ReadFile(absolute)
	if err != nil {
		return parser.Program{}, err
	}
	ts, err := lexer.Scan(string(src))
	if err != nil {
		return parser.Program{}, fmt.Errorf("%s: %w", path, err)
	}
	program, err := parser.Parse(ts)
	if err != nil {
		return parser.Program{}, fmt.Errorf("%s: %w", path, err)
	}
	active[absolute] = true
	defer delete(active, absolute)
	merged := parser.Program{}
	for _, imported := range program.Imports {
		importedPath := imported
		if !filepath.IsAbs(importedPath) {
			importedPath = filepath.Join(filepath.Dir(absolute), importedPath)
		}
		dependency, err := loadProgram(importedPath, active, cache)
		if err != nil {
			return parser.Program{}, err
		}
		merged.Functions = append(merged.Functions, dependency.Functions...)
		merged.Types = append(merged.Types, dependency.Types...)
		merged.Capabilities = append(merged.Capabilities, dependency.Capabilities...)
	}
	merged.Functions = append(merged.Functions, program.Functions...)
	merged.Types = append(merged.Types, program.Types...)
	merged.Capabilities = append(merged.Capabilities, program.Capabilities...)
	for _, function := range merged.Functions {
		if function.Name == "main" {
			merged.Body = function.Body
		}
	}
	cache[absolute] = merged
	return merged, nil
}
