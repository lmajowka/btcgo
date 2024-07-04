/**
 * BTCGO
 *
 * Modulo : base functions
 */

package utils

import (
	"math/big"
)

// base58Encode encodes a byte slice to a base58-encoded string.
var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

var b58 = [256]byte{
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 0, 1, 2, 3, 4, 5, 6,
	7, 8, 255, 255, 255, 255, 255, 255,
	255, 9, 10, 11, 12, 13, 14, 15,
	16, 255, 17, 18, 19, 20, 21, 255,
	22, 23, 24, 25, 26, 27, 28, 29,
	30, 31, 32, 255, 255, 255, 255, 255,
	255, 33, 34, 35, 36, 37, 38, 39,
	40, 41, 42, 43, 255, 44, 45, 46,
	47, 48, 49, 50, 51, 52, 53, 54,
	55, 56, 57, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
}

var bigRadix10 = big.NewInt(58 * 58 * 58 * 58 * 58 * 58 * 58 * 58 * 58 * 58) // 58^10
var bigRadix = [...]*big.Int{
	big.NewInt(0),
	big.NewInt(58),
	big.NewInt(58 * 58),
	big.NewInt(58 * 58 * 58),
	big.NewInt(58 * 58 * 58 * 58),
	big.NewInt(58 * 58 * 58 * 58 * 58),
	big.NewInt(58 * 58 * 58 * 58 * 58 * 58),
	big.NewInt(58 * 58 * 58 * 58 * 58 * 58 * 58),
	big.NewInt(58 * 58 * 58 * 58 * 58 * 58 * 58 * 58),
	big.NewInt(58 * 58 * 58 * 58 * 58 * 58 * 58 * 58 * 58),
	bigRadix10,
}

func Encode(input []byte) string {
	var result []byte
	x := new(big.Int).SetBytes(input)

	base := big.NewInt(int64(len(base58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	// Reverse the result
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	// Add leading zeroes
	for _, b := range input {
		if b != 0 {
			break
		}
		result = append([]byte{base58Alphabet[0]}, result...)
	}

	return string(result)
}

func Decode(b string) []byte {
	answer := big.NewInt(0)
	scratch := new(big.Int)

	for t := b; len(t) > 0; {
		n := len(t)
		if n > 10 {
			n = 10
		}

		total := uint64(0)
		for _, v := range t[:n] {
			if v > 255 {
				return []byte("")
			}

			tmp := b58[v]
			if tmp == 255 {
				return []byte("")
			}
			total = total*58 + uint64(tmp)
		}

		answer.Mul(answer, bigRadix[n])
		scratch.SetUint64(total)
		answer.Add(answer, scratch)

		t = t[n:]
	}

	tmpval := answer.Bytes()

	var numZeros int
	for numZeros = 0; numZeros < len(b); numZeros++ {
		if b[numZeros] != '1' {
			break
		}
	}
	flen := numZeros + len(tmpval)
	val := make([]byte, flen)
	copy(val[numZeros:], tmpval)

	return val
}
