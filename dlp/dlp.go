package dlp

import (
	"fmt"
	"strings"

	"github.com/David2024patton/itak-shield/scanner"
)

// Action defines what happens when a PII type is detected.
type Action string

const (
	ActionRedact Action = "redact" // default: replace with token
	ActionBlock  Action = "block"  // reject the entire request
)

// Policy maps PII type names to actions.
type Policy struct {
	rules map[string]Action // PIIType name -> action
}

// Result holds the DLP evaluation result.
type Result struct {
	Blocked      bool     // true if request should be rejected
	BlockedTypes []string // which PII types triggered the block
	Message      string   // human-readable block reason
}

// New creates a DLP policy engine from a map of type -> action strings.
// Unknown actions default to "redact".
func New(rules map[string]string) *Policy {
	if len(rules) == 0 {
		return nil
	}

	p := &Policy{
		rules: make(map[string]Action, len(rules)),
	}
	for piiType, action := range rules {
		switch strings.ToLower(action) {
		case "block":
			p.rules[strings.ToUpper(piiType)] = ActionBlock
		default:
			p.rules[strings.ToUpper(piiType)] = ActionRedact
		}
	}
	return p
}

// Evaluate checks scanner matches against DLP policies.
// Returns a Result indicating whether to block or proceed with redaction.
func (p *Policy) Evaluate(matches []scanner.Match) *Result {
	if p == nil {
		return &Result{Blocked: false}
	}

	var blockedTypes []string

	for _, m := range matches {
		typeName := strings.ToUpper(string(m.Type))
		action, ok := p.rules[typeName]
		if ok && action == ActionBlock {
			blockedTypes = append(blockedTypes, typeName)
		}
	}

	if len(blockedTypes) > 0 {
		// Deduplicate.
		seen := make(map[string]bool)
		var unique []string
		for _, t := range blockedTypes {
			if !seen[t] {
				seen[t] = true
				unique = append(unique, t)
			}
		}
		return &Result{
			Blocked:      true,
			BlockedTypes: unique,
			Message:      fmt.Sprintf("Request blocked by DLP policy: contains %s", strings.Join(unique, ", ")),
		}
	}

	return &Result{Blocked: false}
}

// GetAction returns the configured action for a PII type.
// Returns ActionRedact if not configured or policy is nil.
func (p *Policy) GetAction(piiType string) Action {
	if p == nil {
		return ActionRedact
	}
	action, ok := p.rules[strings.ToUpper(piiType)]
	if !ok {
		return ActionRedact
	}
	return action
}
