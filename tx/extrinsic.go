package tx

import (
	"encoding/binary"
	"errors"
	"fmt"
	codec "github.com/JFJun/chainX-go/codes"
	"github.com/JFJun/chainX-go/ss58"
	"github.com/JFJun/chainX-go/util"
	"strconv"
)

type ChainXExtrinsic struct {
	offset        int
	rawData       []byte
	data          []byte
	compactLength int
	CallIndex     string
	From          string
	To            string
	Signature     string
	Nonce         uint64
	Era           string
	Acceleration  int
	Token         string
	Amount        uint64
	Memo          string
	Timestamp     int64
}

func NewChainXExtrinsic(data []byte) *ChainXExtrinsic {
	ce := new(ChainXExtrinsic)
	ce.data = data
	ce.offset = 0
	return ce
}
func (ce *ChainXExtrinsic) getNextBytes(length int) []byte {
	if ce.offset+length > len(ce.data) {
		data := ce.data[ce.offset:]
		ce.offset = len(ce.data)
		return data
	}
	data := ce.data[ce.offset : ce.offset+length]
	ce.offset = ce.offset + length
	return data
}

func (ce *ChainXExtrinsic) processCompactBytes() []byte {
	// 初始化
	ce.compactLength = 0
	compactByte := ce.getNextBytes(1)
	var byteMod = 0
	if len(compactByte) != 0 {
		byteMod = int(compactByte[0]) % 4
	}
	if byteMod == 0 {
		ce.compactLength = 1
	} else if byteMod == 1 {
		ce.compactLength = 2
	} else if byteMod == 2 {
		ce.compactLength = 4
	} else {
		ce.compactLength = 5 + ((int(compactByte[0]) - 3) / 4)
	}
	var CompactBytes []byte
	if ce.compactLength == 1 {
		CompactBytes = compactByte
	} else if util.IntInSlice(ce.compactLength, []int{2, 4}) {
		CompactBytes = append(compactByte[:], ce.getNextBytes(ce.compactLength - 1)[:]...)
	} else {
		CompactBytes = ce.getNextBytes(ce.compactLength - 1)
	}
	return CompactBytes
}

func (ce *ChainXExtrinsic) ParseChainXExtrinsic() error {
	compactBytes := ce.processCompactBytes()

	length, err := codec.DecodeCompact32(compactBytes)

	if err != nil {
		return fmt.Errorf("parse extrinsic length to  compact32 error,err=%v", err)
	}

	l := int(length)
	if l != len(ce.data[ce.compactLength:]) {
		return errors.New("extrinsic length is not equal")
	}

	versionInfo := util.BytesToHex(ce.getNextBytes(1))

	containsTx := util.HexToU256(versionInfo).Int64() >= 80
	if versionInfo == "01" || versionInfo == "81" {
		if containsTx {
			//解析from地址
			var (
				err error
			)
			ce.From, err = ce.parseAddress()

			if err != nil {
				return fmt.Errorf("get from address error,err=%v", err)
			}
			//解析签名
			ce.Signature = util.BytesToHex(ce.getNextBytes(64))

			//解析nonce
			nonceCompactBytes := ce.processCompactBytes()

			nonce, err := codec.DecodeCompact32(nonceCompactBytes)
			if err != nil {
				return fmt.Errorf("parse nonce length to  compact32 error,err=%v", err)
			}
			ce.Nonce = uint64(nonce)

			//解析 era
			era := util.BytesToHex(ce.getNextBytes(1))
			if era != "00" {
				return errors.New("era is not equal '00'")
			}
			// 	解析 acceleration

			acc, err := codec.DecodeCompact32(ce.processCompactBytes())
			if err != nil {
				return fmt.Errorf("decode acceleration error,err=%v", err)
			}
			ce.Acceleration = int(acc)

		}

		ce.CallIndex = util.BytesToHex(ce.getNextBytes(2))
	} else {
		return fmt.Errorf("Extrinsic version %s is not support", versionInfo)
	}
	if ce.CallIndex != "" {
		return ce.parseCallIndex()

	}
	return nil
}

func (ce *ChainXExtrinsic) parseAddress() (string, error) {
	al := ce.getNextBytes(1)
	AccountLength := util.BytesToHex(al)
	var (
		address string
		err     error
	)
	if AccountLength == "ff" {
		address, err = ss58.EncodeByPubHex(util.BytesToHex(ce.getNextBytes(32)), ss58.ChainXPrefix)
	} else {
		num, _ := strconv.ParseUint(AccountLength, 16, 32)
		address = fmt.Sprintf("%d", num)
		//address, err = ss58.EncodeByPubHex(util.BytesToHex(append(al, ce.getNextBytes(31)...)), ss58.ChainXPrefix)
	}
	if err != nil {
		return "", fmt.Errorf("parse address error,err=%v", err)
	}
	return address, nil
}

func (ce *ChainXExtrinsic) parseCallIndex() error {
	var (
		err error
	)
	if ce.CallIndex == CallIdTransfer {
		//	表示为 一笔交易
		ce.To, err = ce.parseAddress()
		if err != nil {
			return fmt.Errorf("parse to address error,err=%v", err)
		}
		//解析token长度
		tl := ce.processCompactBytes()
		tLen, err := codec.DecodeCompact32(tl)
		if err != nil {
			return fmt.Errorf("parse token length error,err=%v", err)
		}
		tLength := int(tLen)
		//解析token

		t := ce.getNextBytes(tLength)

		ce.Token = string(t)
		//解析Amount
		ab := ce.getNextBytes(8)
		ce.Amount = binary.LittleEndian.Uint64(ab)
		//	解析memo
		ml := ce.processCompactBytes()
		if ml == nil || len(ml) == 0 {
			//  memo is null
			ce.Memo = ""
		} else {
			mLen, err := codec.DecodeCompact32(ml)
			if err != nil {
				return fmt.Errorf("decode memo length error,err=%v", err)
			}
			mLength := int(mLen)
			if mLength == 0 {
				ce.Memo = ""
			} else {
				ce.Memo = string(ce.getNextBytes(mLength))
			}

		}

	} else if ce.CallIndex == CallIdTimestamp {
		t := util.BytesToHex(ce.getNextBytes(1))
		if t == "03" {
			//解析时间戳
			tu, err := codec.DecodeUint64(ce.getNextBytes(4))
			if err != nil {
				return fmt.Errorf("parse timestamp error,err=%v", err)
			}
			ce.Timestamp = int64(tu)
		}
	} else if ce.CallIndex == CallIdProduce {
		return nil
	}
	return nil
}
