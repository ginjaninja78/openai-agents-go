package pod

import (
	"fmt"
)

// TaskNode represents a single, atomic unit of work in the Swarm.
type TaskNode struct {
	ID           string
	Action       string
	DomainLease  string // e.g., "/frontend" or "/auth"
	Dependencies []string
}

// LSystemEngine holds the hardcoded grammar rules.
type LSystemEngine struct {
	rules map[string][]TaskNode
}

// NewLSystemEngine initializes the deterministic grammars.
func NewLSystemEngine() *LSystemEngine {
	engine := &LSystemEngine{
		rules: make(map[string][]TaskNode),
	}

	// Example: The deterministic grammar for building an API endpoint.
	// The LLM only has to say "CreateAPIEndpoint". It cannot hallucinate these steps.
	engine.rules["CreateAPIEndpoint"] = []TaskNode{
		{ID: "step1", Action: "Define_Structs", DomainLease: "/api"},
		{ID: "step2", Action: "Write_Handler", DomainLease: "/api", Dependencies: []string{"step1"}},
		{ID: "step3", Action: "Write_Tests", DomainLease: "/api", Dependencies: []string{"step2"}},
		{ID: "step4", Action: "Sandbox_2PC_Verify", DomainLease: "/api", Dependencies: []string{"step3"}},
	}
	
	return engine
}

// Expand takes the LLM's chosen rule and generates the execution DAG.
func (e *LSystemEngine) Expand(ruleName string, depth int) ([]TaskNode, error) {
	nodes, exists := e.rules[ruleName]
	if !exists {
		return nil, fmt.Errorf("fatal: LLM attempted to use invalid grammar rule: %s", ruleName)
	}

	// In a full implementation, `depth` would recursively expand nodes that 
	// contain sub-rules, generating a massive deterministic DAG instantly.
	return nodes, nil
}