package llm

import (
	"fmt"
	"sync"
	"time"
)

// ModelPool manages the multi-model fallback chain and circuit breaker
type ModelPool struct {
	primaryModel     string
	candidates       []string
	cooldownDuration time.Duration
	cooldowns        map[string]time.Time
	metadata         map[string]ModelMetadata
	mu               sync.RWMutex
}

// NewModelPool creates a new ModelPool with the given primary model and fallback candidates
func NewModelPool(primary string, fallbacks []string, cooldown time.Duration) *ModelPool {
	if cooldown <= 0 {
		cooldown = 60 * time.Second
	}

	seen := make(map[string]bool)
	var ordered []string

	if primary != "" {
		ordered = append(ordered, primary)
		seen[primary] = true
	}

	for _, m := range fallbacks {
		if m != "" && !seen[m] {
			ordered = append(ordered, m)
			seen[m] = true
		}
	}

	return &ModelPool{
		primaryModel:     primary,
		candidates:       ordered,
		cooldownDuration: cooldown,
		cooldowns:        make(map[string]time.Time),
		metadata:         make(map[string]ModelMetadata),
	}
}

// UpdateAvailableModels updates known capabilities and availability from /v1/models
func (p *ModelPool) UpdateAvailableModels(models []ModelMetadata) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, m := range models {
		p.metadata[m.ID] = m
	}
}

// ReportFailure flags a model as failed and places it into cooldown if rate-limited (429) or overloaded (5xx)
func (p *ModelPool) ReportFailure(modelID string, statusCode int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Only apply cooldown on rate limits, quota limits, or server downtime
	if statusCode == 429 || statusCode == 503 || statusCode == 502 || statusCode == 500 {
		p.cooldowns[modelID] = time.Now().Add(p.cooldownDuration)
	}
}

// ReportSuccess acknowledges successful response; resets cooldown if any
func (p *ModelPool) ReportSuccess(modelID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.cooldowns, modelID)
}

// IsOnCooldown returns whether the specified model is currently cooling down
func (p *ModelPool) IsOnCooldown(modelID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	until, exists := p.cooldowns[modelID]
	if !exists {
		return false
	}
	return time.Now().Before(until)
}

// GetCooldownRemaining returns how long until cooldown expires (0 if not in cooldown)
func (p *ModelPool) GetCooldownRemaining(modelID string) time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	until, exists := p.cooldowns[modelID]
	if !exists {
		return 0
	}
	rem := time.Until(until)
	if rem < 0 {
		return 0
	}
	return rem
}

// GetNextAvailableModel finds the best candidate model satisfying requirements and not in cooldown
func (p *ModelPool) GetNextAvailableModel(requiresTools bool, minContext int) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	now := time.Now()
	var earliestCooldownTime time.Time
	var earliestCooldownModel string

	for _, modelID := range p.candidates {
		// 1. Check metadata availability if known
		if meta, ok := p.metadata[modelID]; ok {
			if !meta.Available {
				continue
			}
			if requiresTools && !meta.SupportsTools() {
				continue
			}
			if minContext > 0 && meta.ContextWindow > 0 && meta.ContextWindow < minContext {
				continue
			}
		}

		// 2. Check cooldown status
		if until, inCooldown := p.cooldowns[modelID]; inCooldown && now.Before(until) {
			if earliestCooldownTime.IsZero() || until.Before(earliestCooldownTime) {
				earliestCooldownTime = until
				earliestCooldownModel = modelID
			}
			continue
		}

		// Found available candidate!
		return modelID, nil
	}

	if earliestCooldownModel != "" {
		return earliestCooldownModel, fmt.Errorf("all eligible models are in cooldown; earliest candidate '%s' ready in %v",
			earliestCooldownModel, time.Until(earliestCooldownTime).Round(time.Second))
	}

	return "", fmt.Errorf("no suitable models available in fallback pool matching requirements (requiresTools=%v, minContext=%d)", requiresTools, minContext)
}

// AllCandidates returns a copy of the candidate list
func (p *ModelPool) AllCandidates() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]string, len(p.candidates))
	copy(out, p.candidates)
	return out
}
