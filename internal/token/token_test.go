package token

import "testing"

func TestGenerateDefaultLengthAndDigits(t *testing.T) {
	got, err := Generate(0)
	if err != nil {
		t.Fatalf("Generate(0) error: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("default token length = %d, want 4", len(got))
	}
	for i := 0; i < len(got); i++ {
		if got[i] < '0' || got[i] > '9' {
			t.Fatalf("token contains non-digit %q at index %d", got[i], i)
		}
	}
}

func TestGenerateCustomLengthDigitsOnly(t *testing.T) {
	const n = 8
	got, err := Generate(n)
	if err != nil {
		t.Fatalf("Generate(%d) error: %v", n, err)
	}
	if len(got) != n {
		t.Fatalf("token length = %d, want %d", len(got), n)
	}
	for i := 0; i < len(got); i++ {
		if got[i] < '0' || got[i] > '9' {
			t.Fatalf("token contains non-digit %q at index %d", got[i], i)
		}
	}
}
