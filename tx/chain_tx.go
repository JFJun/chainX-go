package tx

import (
	"encoding/hex"
	"errors"
	codec "github.com/JFJun/chainX-go/codes"
	"github.com/JFJun/chainX-go/ss58"
	"strings"

	"golang.org/x/crypto/ed25519"
)

/*
新增chainX的交易
*/

const (
	CallIdTransfer  = "0803"
	CallIdTimestamp = "0100"
	CallIdProduce   = "0600"
)

type ChainXTransaction struct {
	SenderPubkey    string `json:"sender_pubkey"`    // from address public key ,0x开头
	RecipientPubkey string `json:"recipient_pubkey"` // to address public key ,0x开头
	Amount          uint64 `json:"amount"`           // 转账金额
	Nonce           uint64 `json:"nonce"`            //nonce值
	Acceleration    uint64 `json:"fee"`              //
	//BlockHeight        uint64 `json:"block_height"`     //最新区块高度
	BlockHash   string `json:"block_hash"`   //最新区块hash
	GenesisHash string `json:"genesis_hash"` //
	//SpecVersion        uint32 `json:"spec_version"`
	//TransactionVersion uint32 `json:"transaction_version"`
	Token  string `json:"token"`
	Memo   string `json:"memo"`
	CallId string `json:"call_id"`
}
type ChainXTransferParams struct {
	From   string
	To     string
	Token  string
	Amount uint64
	Nonce  uint64
	Memo   string
}

func CreateChainXTransaction(params *ChainXTransferParams) *ChainXTransaction {
	tx := ChainXTransaction{
		SenderPubkey:    AddressToPublicKey(params.From),
		RecipientPubkey: AddressToPublicKey(params.To),
		Amount:          params.Amount,
		Nonce:           params.Nonce,
		Token:           params.Token,
		Memo:            params.Memo,
	}
	//Acceleration 默认设置为 1
	tx.Acceleration = 1
	return &tx
}

func (t *ChainXTransaction) SetBlockHashAndCallId(blockHash, callId string) {
	t.BlockHash = Remove0X(blockHash)
	t.CallId = callId
}

func (t *ChainXTransaction) CreatSignData() (string, error) {
	tp, err := t.newSignData()
	if err != nil {
		return "", err
	}
	return tp.Encode(), nil
}

func (t *ChainXTransaction) Sign(private, message string) (string, error) {
	message = Remove0X(message)
	messageBytes, err := hex.DecodeString(message)
	if err != nil {
		return "", err
	}
	private = Remove0X(private)
	priv, err1 := hex.DecodeString(private)
	if err1 != nil {
		return "", err1
	}
	//使用ed25519签名
	privKey := ed25519.NewKeyFromSeed(priv)
	sig := ed25519.Sign(privKey, messageBytes)
	if len(sig) != 64 {
		return "", errors.New("sign fail,sig length is not equal 64")
	}
	return hex.EncodeToString(sig), nil
}

func (t *ChainXTransaction) CombineChainXtx(signature string) (string, error) {
	signed := make([]byte, 0)
	//签名版本号
	signed = append(signed, SigningChainXBit)
	signed = append(signed, 0xff)
	//from地址
	from, err := hex.DecodeString(t.SenderPubkey)

	if err != nil || len(from) != 32 {
		return "", nil
	}
	signed = append(signed, from...)

	//签名数据
	//signed = append(signed, 0x01) // ed25519: 0x00 sr25519: 0x01
	signature = Remove0X(signature)
	sig, err := hex.DecodeString(signature)
	if err != nil || len(sig) != 64 {
		return "", nil
	}
	signed = append(signed, sig...)

	//nonce
	if t.Nonce == 0 {
		signed = append(signed, []byte{0}...)
	} else {
		nonce, err := codec.Encode(Compact_U32, uint64(t.Nonce))
		if err != nil {
			return "", err
		}
		nonceBytes, _ := hex.DecodeString(nonce)
		signed = append(signed, nonceBytes...)
	}
	// era
	signed = append(signed, []byte{0x00}...)

	//acceleration
	acceleration, err := codec.Encode(Compact_U32, t.Acceleration)
	if err != nil {
		return "", err
	}
	accBytes, _ := hex.DecodeString(acceleration)
	signed = append(signed, accBytes...)

	method, err := NewChainXMethodTransfer(t.RecipientPubkey, t.Token, t.Memo, t.Amount)
	if err != nil {
		return "", err
	}

	methodBytes := method.Encode(t.CallId)

	signed = append(signed, methodBytes...)

	length, err := codec.Encode(Compact_U32, uint64(len(signed)))

	if err != nil {
		return "", err
	}
	lengthBytes, _ := hex.DecodeString(length)

	lengthBytes[0] += 1

	return "0x" + hex.EncodeToString(lengthBytes) + hex.EncodeToString(signed), nil
}

func (t *ChainXTransaction) newSignData() (*ChainXSignaturePayload, error) {
	tp := new(ChainXSignaturePayload)
	if t.Nonce == 0 {
		tp.Nonce = []byte{0}
	} else {
		nonce, err := codec.Encode(Compact_U32, uint64(t.Nonce))
		if err != nil {
			return nil, err
		}
		tp.Nonce, _ = hex.DecodeString(nonce)
	}
	//method
	method, err := NewChainXMethodTransfer(t.RecipientPubkey, t.Token, t.Memo, t.Amount)
	if err != nil {
		return nil, err
	}
	tp.Method = method.Encode(t.CallId)

	//era
	tp.Era = []byte{0x00}

	//blockHash
	block, err := hex.DecodeString(Remove0X(t.BlockHash))
	if err != nil || len(block) != 32 {
		return nil, errors.New("invalid block hash")
	}
	tp.BlockHash = block
	// acceleration
	acceleration, err := codec.Encode(Compact_U32, t.Acceleration)
	if err != nil {
		return nil, err
	}
	tp.Acceleration, _ = hex.DecodeString(acceleration)
	return tp, nil
}
func AddressToPublicKey(address string) string {
	if address == "" {
		return ""
	}
	pub, err := ss58.DecodeToPub(address)

	if err != nil {
		return ""
	}
	if len(pub) != 32 {
		return ""
	}
	pubHex := hex.EncodeToString(pub)
	return pubHex
}

func Remove0X(hexData string) string {
	if strings.HasPrefix(hexData, "0x") {
		return hexData[2:]
	}
	return hexData
}
