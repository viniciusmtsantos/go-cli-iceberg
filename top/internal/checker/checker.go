package checker

import (
	"os/exec"
	"strings"
	"time"
)

// Tool defines a CLI tool to be checked.
type Tool struct {
	Name    string
	Command string
	Args    []string
}

// Result holds the outcome of checking a single tool.
type Result struct {
	Tool    string
	Version string
	Found   bool
	Latency time.Duration
}

// DefaultTools is the list goprobe checks by default.
var DefaultTools = []Tool{
	{Name: "Go", Command: "go", Args: []string{"version"}},
	{Name: "Git", Command: "git", Args: []string{"--version"}},
	{Name: "Docker", Command: "docker", Args: []string{"--version"}},
	{Name: "kubectl", Command: "kubectl", Args: []string{"version", "--client"}},
}

// Check verifies whether a tool is installed and captures its version string.
func Check(tool Tool) Result {
	start := time.Now()
	out, err := exec.Command(tool.Command, tool.Args...).Output()
	latency := time.Since(start)

	if err != nil {
		return Result{Tool: tool.Name, Found: false, Latency: latency}
	}

	version := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	return Result{
		Tool:    tool.Name,
		Version: version,
		Found:   true,
		Latency: latency,
	}
}

// CheckAll runs Check for each tool in the list sequentially.
func CheckAll(tools []Tool) []Result {
	results := make([]Result, len(tools))
	for i, tool := range tools {
		results[i] = Check(tool)
	}
	return results
}
