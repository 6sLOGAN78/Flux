// Package base62 implements a high-performance Base62 encoder and decoder for URL shorteners.
package base62

import (
	"errors"
	"fmt"
	"strings"
)

const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base     = 62
)

var (
	decodeMap [256]int8
	ErrInvalidCharacter = errors.New("invalid character in base62 string")
	ErrOverflow         = errors.New("base62 decoding integer overflow")
)

func init() {
	for i := 0; i < 256; i++ {
		decodeMap[i] = -1
	}
	for i := 0; i < len(alphabet); i++ {
		decodeMap[alphabet[i]] = int8(i)
	}
}

// Encode converts a 64-bit unsigned integer to a Base62 string representation.
func Encode(num uint64) string {
	if num == 0 {
		return "0"
	}

	var buf [11]byte
	i := len(buf)

	for num > 0 {
		i--
		buf[i] = alphabet[num%base]
		num /= base
	}

	return string(buf[i:])
}

// EncodePadded converts a 64-bit unsigned integer to a Base62 string padded with leading zeros to at least minLength.
func EncodePadded(num uint64, minLength int) string {
	encoded := Encode(num)
	if len(encoded) >= minLength {
		return encoded
	}
	padding := strings.Repeat("0", minLength-len(encoded))
	return padding + encoded
}

// Decode converts a Base62 encoded string back to a 64-bit unsigned integer.
func Decode(str string) (uint64, error) {
	if len(str) == 0 {
		return 0, errors.New("empty string cannot be decoded")
	}

	var result uint64
	for i := 0; i < len(str); i++ {
		char := str[i]
		val := decodeMap[char]
		if val < 0 {
			return 0, fmt.Errorf("%w: '%c'", ErrInvalidCharacter, char)
		}

		// Check overflow before multiplication
		if result > (mathMaxUint64-uint64(val))/base {
			return 0, ErrOverflow
		}

		result = result*base + uint64(val)
	}

	return result, nil
}

const mathMaxUint64 = 18446744073709551615
