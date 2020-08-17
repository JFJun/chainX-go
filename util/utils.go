package util

import (
	"encoding/hex"
	"errors"
	"github.com/JFJun/chainX-go/xxhash"
	"math/big"

	"golang.org/x/crypto/blake2b"
	"hash"
	"strings"
)

func AppendBytes(data1, data2 []byte) []byte {
	if data2 == nil {
		return data1
	}
	return append(data1, data2...)
}

func SelectHash(method string) (hash.Hash, error) {
	switch method {
	case "Twox128":
		return xxhash.New128(nil), nil
	case "Blake2_256":
		return blake2b.New256(nil)
	case "Blake2_128":
		return blake2b.New(16, nil)
	case "Blake2_128Concat":
		return blake2b.New(16, nil)
	case "Twox64Concat":
		return xxhash.New64(nil), nil
	case "Identity":
		return nil, nil
	default:
		return nil, errors.New("unknown hash method")

	}

}

func RemoveHex0x(hexStr string) string {
	if strings.HasPrefix(hexStr, "0x") {
		return hexStr[2:]
	}
	return hexStr
}

func IntInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func BytesToHex(b []byte) string {
	c := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(c, b)
	return string(c)
}
func HexToU256(v string) *big.Int {
	v = strings.TrimPrefix(v, "0x")
	bn := new(big.Int)
	n, _ := bn.SetString(v, 16)
	return n
}
