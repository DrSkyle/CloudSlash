package heuristics

import (
	"context"
	"fmt"
	"sync"

	"github.com/DrSkyle/cloudslash/internal/graph"
)

// WasteConfidence represents the probability (0.0-1.0) that a resource is waste.
type WasteConfidence float64

// HeuristicResult contains the analysis outcome.
type HeuristicResult struct {
	IsWaste    bool
	Confidence WasteConfidence
	RiskScore  int // 0-100 impact of deletion (Safety)
	Reason     string
}

// WeightedHeuristic defines a sophisticated analyzer.
type WeightedHeuristic interface {
	Name() string
	Run(ctx context.Context, g *graph.Graph) error
}

// Engine orchestrates the heuristic analysis.
type Engine struct {
	heuristics []WeightedHeuristic
}

// NewEngine creates a new Heuristic Engine.
func NewEngine() *Engine {
	return &Engine{
		heuristics: []WeightedHeuristic{},
	}
}

// Register adds a heuristic to the engine.
func (e *Engine) Register(h WeightedHeuristic) {
	e.heuristics = append(e.heuristics, h)
}

// Run executes all registered heuristics concurrently.
func (e *Engine) Run(ctx context.Context, g *graph.Graph) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(e.heuristics))

	for _, h := range e.heuristics {
		wg.Add(1)
		go func(h WeightedHeuristic) {
			defer wg.Done()
			if err := h.Run(ctx, g); err != nil {
				errChan <- fmt.Errorf("heuristic %s failed: %w", h.Name(), err)
			}
		}(h)
	}

	wg.Wait()
	close(errChan)

	// Collect errors (fairly basic error handling for now)
	for err := range errChan {
		// Log error but don't stop entire engine?
		// For now, let's just print to stdout or return first error
		return err
	}

	return nil
}
