package tx

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	codec "github.com/JFJun/chainX-go/codes"
)

const (
	SigningChainXBit = byte(0x81)
	Compact_U32      = "Compact<u32>"
)

type ChainXMethodTransfer struct {
	DestPubkey  []byte
	TokenLength []byte
	Token       []byte
	Amount      []byte
	MemoLength  []byte
	Memo        []byte
}

func NewChainXMethodTransfer(pubkey, token, memo string, amount uint64) (*ChainXMethodTransfer, error) {
	//to地址公钥
	pubBytes, err := hex.DecodeString(pubkey)
	if err != nil || len(pubBytes) != 32 {
		return nil, errors.New("invalid dest public key")
	}

	//编码 使用的是小端字节
	if amount == 0 {
		return nil, errors.New("zero amount")
	}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, amount)

	//编码 token
	tk, err := codec.Encode("string", token)
	if err != nil {
		return nil, fmt.Errorf("encode token error,err=%v", err)
	}
	tokenBytes, _ := hex.DecodeString(tk)

	//编码 token长度 一般token的长度不会大于256*4,所以使用 int直接转换为[]byte
	tokenLength := []byte{byte(len(tokenBytes) * 4)}

	//编码memo
	var (
		memoBytes    []byte
		memoLenBytes []byte
	)
	if memo != "" {
		mm, err := codec.Encode("string", memo)
		if err != nil {
			return nil, fmt.Errorf("encode token error,err=%v", err)
		}

		//编码memo的长度
		memoBytes, _ = hex.DecodeString(mm)
		mLen := len(memoBytes) * 4
		if mLen < 256 {
			memoLenBytes = []byte{byte(mLen)}
		} else {
			lenBuf := make([]byte, 2)
			binary.LittleEndian.PutUint16(lenBuf, uint16(mLen))
			memoLenBytes = lenBuf
		}
	} else {
		memoLenBytes = []byte{0}
	}
	return &ChainXMethodTransfer{
		DestPubkey:  pubBytes,
		Token:       tokenBytes,
		Memo:        memoBytes,
		Amount:      buf,
		TokenLength: tokenLength,
		MemoLength:  memoLenBytes,
	}, nil
}

func (mt *ChainXMethodTransfer) Encode(callId string) []byte {
	ret, _ := hex.DecodeString(callId)
	ret = append(ret, 0xff)
	ret = append(ret, mt.DestPubkey...)
	ret = append(ret, mt.TokenLength...)
	ret = append(ret, mt.Token...)
	ret = append(ret, mt.Amount...)
	ret = append(ret, mt.MemoLength...)
	ret = append(ret, mt.Memo...)
	return ret

}
