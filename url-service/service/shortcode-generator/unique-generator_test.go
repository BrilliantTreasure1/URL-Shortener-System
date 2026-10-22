package shortcodegenerator

import (
	"strings"
	"testing"
)

func TestEncodeBase62(t *testing.T) {
	tests := []struct {
		name string
		n    int64
		want string
	}{
		{"zero", 0, "0"},
		{"one", 1, "1"},
		{"last upper case", 35, "Z"},
		{"first lower case", 36, "a"},
		{"last single digit", 61, "z"},
		{"first two digits", 62, "10"},
		{"next two digits", 63, "11"},
		{"larger value", 124, "20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encodeBase62(tt.n); got != tt.want {
				t.Errorf("encodeBase62(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestGenerateUniqueInvalidInputs(t *testing.T) {
	tests := []struct {
		name         string
		sequence     int64
		randomLength int
		wantErr      string
	}{
		{"zero sequence", 0, 5, "invalid sequence"},
		{"negative sequence", -5, 5, "invalid sequence"},
		{"zero random length", 1, 0, "invalid random length"},
		{"negative random length", 1, -1, "invalid random length"},
	}

	g := NewUniqueGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := g.GenerateUnique(tt.sequence, tt.randomLength)

			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
			}
			if got != "" {
				t.Errorf("expected an empty result on error, got %q", got)
			}
		})
	}
}

func TestGenerateRandomLetters(t *testing.T) {
	for _, length := range []int{1, 5, 20} {
		letters, err := generateRandomLetters(length)
		if err != nil {
			t.Fatalf("generateRandomLetters(%d) returned an error: %v", length, err)
		}

		if len(letters) != length {
			t.Fatalf("expected %d characters, got %d", length, len(letters))
		}

		for _, ch := range letters {
			if !strings.ContainsRune(base62Alphabet, ch) {
				t.Errorf("character %q is not part of the base62 alphabet", ch)
			}
		}
	}
}

func TestGenerateUniqueSuffix(t *testing.T) {
	g := NewUniqueGenerator()

	const (
		sequence     = int64(1)
		randomLength = 5
	)

	code, err := g.GenerateUnique(sequence, randomLength)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	counter := encodeBase62(sequence)
	if !strings.HasPrefix(code, counter) {
		t.Fatalf("expected code %q to start with counter part %q", code, counter)
	}

	suffix := code[len(counter):]
	if len(suffix) != randomLength {
		t.Fatalf("expected suffix length %d, got %d (%q)", randomLength, len(suffix), suffix)
	}

	for _, ch := range suffix {
		if !strings.ContainsRune(base62Alphabet, ch) {
			t.Errorf("suffix character %q is not part of the base62 alphabet", ch)
		}
	}
}

// Randomness alone cannot prove uniqueness, so this test bounds the collision
// probability instead of asserting a mathematical guarantee.
//
// With a suffix of length 8 over a 62-character alphabet the space is 62^8 ~= 2.18e14.
// Accounting for the small modulo bias in generateRandomLetters (byte % 62), the
// per-pair collision probability of two suffixes is ~4.8e-15. For n = 10,000 draws
// the expected number of collisions is
//
//	C(n, 2) * p ~= 4.9995e7 * 4.8e-15 ~= 2.4e-7
//
// so a spurious failure is extremely unlikely (~1 in 4 million runs) and a failure
// here would almost certainly indicate a real defect in the generator.
func TestGenerateUniqueUniqueness(t *testing.T) {
	const (
		iterations   = 10_000
		sequence     = int64(42)
		randomLength = 8
	)

	g := NewUniqueGenerator()

	seen := make(map[string]struct{}, iterations)
	for i := 0; i < iterations; i++ {
		code, err := g.GenerateUnique(sequence, randomLength)
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}

		if _, duplicated := seen[code]; duplicated {
			t.Fatalf("duplicate short code generated at iteration %d: %q", i, code)
		}
		seen[code] = struct{}{}
	}

	if len(seen) != iterations {
		t.Fatalf("expected %d unique codes, got %d", iterations, len(seen))
	}
}
