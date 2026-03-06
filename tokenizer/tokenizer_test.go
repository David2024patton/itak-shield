package tokenizer

import (
	"testing"

	"github.com/David2024patton/itak-shield/scanner"
)

func TestRedactAndRestore(t *testing.T) {
	sc := scanner.New()
	tm := New()

	text := "Contact john@test.com about SSN 123-45-6789"
	matches := sc.Scan(text)
	redacted, count := tm.Redact(text, matches)

	if count < 2 {
		t.Fatalf("expected at least 2 redactions, got %d", count)
	}

	// Verify PII is gone.
	if contains(redacted, "john@test.com") {
		t.Error("email still present in redacted text")
	}
	if contains(redacted, "123-45-6789") {
		t.Error("SSN still present in redacted text")
	}

	// Verify tokens are present.
	if !contains(redacted, "[EMAIL_1]") {
		t.Error("missing [EMAIL_1] token")
	}
	if !contains(redacted, "[SSN_1]") {
		t.Error("missing [SSN_1] token")
	}

	// Restore.
	restored := tm.Restore(redacted)
	if restored != text {
		t.Errorf("restore failed:\n  got:  %q\n  want: %q", restored, text)
	}

	t.Logf("Redact+Restore: PASSED (redacted %d items)", count)
}

func TestConsistentTokens(t *testing.T) {
	tm := New()

	// Same value should get the same token.
	t1 := tm.GetOrCreate(scanner.PIIEmail, "a@b.com")
	t2 := tm.GetOrCreate(scanner.PIIEmail, "a@b.com")
	if t1 != t2 {
		t.Errorf("inconsistent tokens: %q vs %q", t1, t2)
	}

	// Different values should get different tokens.
	t3 := tm.GetOrCreate(scanner.PIIEmail, "c@d.com")
	if t1 == t3 {
		t.Errorf("same token for different values: %q", t1)
	}

	t.Logf("Consistent tokens: PASSED")
}

func TestReset(t *testing.T) {
	tm := New()
	tm.GetOrCreate(scanner.PIIEmail, "a@b.com")
	tm.Reset()

	stats := tm.Stats()
	if len(stats) != 0 {
		t.Errorf("expected empty stats after reset, got %d", len(stats))
	}

	// After reset, counters restart.
	token := tm.GetOrCreate(scanner.PIIEmail, "x@y.com")
	if token != "[EMAIL_1]" {
		t.Errorf("expected [EMAIL_1] after reset, got %q", token)
	}
	t.Logf("Reset: PASSED")
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && searchString(haystack, needle)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
