package spend

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Pricing defines cost per 1 million tokens for a model tier.
type Pricing struct {
	Input  float64 `yaml:"input" json:"input"`   // USD per 1M input tokens
	Output float64 `yaml:"output" json:"output"` // USD per 1M output tokens
}

// UserSpend tracks cumulative token usage and estimated cost for one user.
type UserSpend struct {
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	EstimatedUSD float64 `json:"estimated_usd"`
}

// Stats holds spend statistics for the analytics dashboard.
type Stats struct {
	ByUser      map[string]*UserSpend `json:"by_user"`
	TotalInput  int64                 `json:"total_input"`
	TotalOutput int64                 `json:"total_output"`
	TotalUSD    float64               `json:"total_usd"`
}

// Tracker monitors token usage and enforces budget limits per group.
type Tracker struct {
	mu      sync.Mutex
	users   map[string]*UserSpend // user -> cumulative spend
	budgets map[string]float64    // group -> max USD
	groups  map[string]string     // user -> group (cached from auth)
	pricing Pricing               // default pricing tier
}

// New creates a spend tracker with the given budgets and pricing.
func New(budgets map[string]float64, pricing Pricing) *Tracker {
	if budgets == nil {
		budgets = make(map[string]float64)
	}
	return &Tracker{
		users:   make(map[string]*UserSpend),
		budgets: budgets,
		groups:  make(map[string]string),
		pricing: pricing,
	}
}

// SetUserGroup caches the user-to-group mapping (called after auth).
func (t *Tracker) SetUserGroup(user, group string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.groups[user] = group
	t.mu.Unlock()
}

// usagePayload is the structure we extract from OpenAI-compatible responses.
type usagePayload struct {
	Usage struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
		TotalTokens      int64 `json:"total_tokens"`
	} `json:"usage"`
}

// TrackResponse parses token usage from an OpenAI-compatible response body
// and adds it to the user's cumulative spend.
func (t *Tracker) TrackResponse(user string, responseBody []byte) {
	if t == nil {
		return
	}

	var payload usagePayload
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return // not an OpenAI-compatible response, skip silently
	}

	input := payload.Usage.PromptTokens
	output := payload.Usage.CompletionTokens
	if input == 0 && output == 0 {
		return // no usage data
	}

	cost := float64(input)/1_000_000*t.pricing.Input +
		float64(output)/1_000_000*t.pricing.Output

	t.mu.Lock()
	defer t.mu.Unlock()

	s, ok := t.users[user]
	if !ok {
		s = &UserSpend{}
		t.users[user] = s
	}

	s.InputTokens += input
	s.OutputTokens += output
	s.EstimatedUSD += cost
}

// CheckBudget returns an error if the user's group has exceeded its budget.
func (t *Tracker) CheckBudget(user string) error {
	if t == nil {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	group, ok := t.groups[user]
	if !ok {
		return nil // no group = no budget
	}

	budget, hasBudget := t.budgets[group]
	if !hasBudget || budget <= 0 {
		return nil // no budget configured
	}

	s, ok := t.users[user]
	if !ok {
		return nil // no spend yet
	}

	if s.EstimatedUSD >= budget {
		return fmt.Errorf("budget exceeded for group %s: $%.2f / $%.2f", group, s.EstimatedUSD, budget)
	}

	return nil
}

// GetStats returns spend statistics for the dashboard.
func (t *Tracker) GetStats() Stats {
	if t == nil {
		return Stats{ByUser: make(map[string]*UserSpend)}
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	stats := Stats{
		ByUser: make(map[string]*UserSpend, len(t.users)),
	}
	for user, s := range t.users {
		clone := *s
		stats.ByUser[user] = &clone
		stats.TotalInput += s.InputTokens
		stats.TotalOutput += s.OutputTokens
		stats.TotalUSD += s.EstimatedUSD
	}
	return stats
}
