package tx

import (
	"encoding/hex"
	"fmt"
)

type ChainXSignaturePayload struct {
	Nonce        []byte
	Method       []byte
	Era          []byte
	BlockHash    []byte
	Acceleration []byte
}

func (t ChainXSignaturePayload) Encode() string {
	payload := make([]byte, 0)
	payload = append(payload, t.Nonce...)
	payload = append(payload, t.Method...)
	payload = append(payload, t.Era...)
	payload = append(payload, t.BlockHash...)
	payload = append(payload, t.Acceleration...)
	//todo if len big than 256 ,use blake2b to hash it

	return hex.EncodeToString(payload)
}
