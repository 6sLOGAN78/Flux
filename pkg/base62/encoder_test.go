package base62_test

import (
	"math"
	"testing"

	"flux/pkg/base62"
)

func TestEncodeAndDecode(t *testing.T) {
	testCases := []struct {
		name     string
		num      uint64
		expected string
	}{
		{"Zero", 0, "0000000"},
		{"One", 1, "0000001"},
		{"Base61", 61, "000000Z"},
		{"Base62", 62, "0000010"},
		{"LargeNumber", 123456789, "008M3k0"},
		{"MaxUint64", math.MaxUint64, "1L7zB3eD097"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := base62.EncodePadded(tc.num, 7)
			if tc.num < 62^7 && len(encoded) < 7 {
				t.Errorf("expected padded length at least 7, got %d", len(encoded))
			}

			decoded, err := base62.Decode(encoded)
			if err != nil {
				t.Fatalf("unexpected error decoding '%s': %v", encoded, err)
			}

			if decoded != tc.num {
				t.Errorf("expected decoded uint64 %d, got %d", tc.num, decoded)
			}
		})
	}
}

func TestDecode_InvalidCharacter(t *testing.T) {
	invalidStrings := []string{
		"0000!00",
		"abc@123",
		"short key",
		"123#456",
	}

	for _, str := range invalidStrings {
		t.Run(str, func(t *testing.T) {
			_, err := base62.Decode(str)
			if err == nil {
				t.Errorf("expected error decoding invalid string '%s', got nil", str)
			}
		})
	}
}

func TestEncode_RoundTrip(t *testing.T) {
	numbers := []uint64{0, 1, 100, 9999, 1234567, 9876543210, 18446744073709551615}

	for _, num := range numbers {
		encoded := base62.Encode(num)
		decoded, err := base62.Decode(encoded)
		if err != nil {
			t.Fatalf("failed to decode '%s': %v", encoded, err)
		}
		if decoded != num {
			t.Fatalf("roundtrip failed for %d: got %d", num, decoded)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = base62.Encode(123456789)
	}
}

func BenchmarkDecode(b *testing.B) {
	encoded := base62.Encode(123456789)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = base62.Decode(encoded)
	}
}
