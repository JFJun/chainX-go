package model

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
)

const (
	XASSETS    = "xassets"
	XFEEMANAGE = "xfee_manager"
)

type ChainXEventData struct {
	Symbol  string
	Token   string
	RawData string
	Amount  int64
}

func ParseChainXEventData(data string) (*ChainXEventData, error) {
	ced := new(ChainXEventData)
	ced.RawData = data
	if strings.HasPrefix(data, XASSETS) {
		ced.Symbol = XASSETS
	} else if strings.HasPrefix(data, XFEEMANAGE) {
		ced.Symbol = XFEEMANAGE
	} else {
		return ced, errors.New("do not support this event parse")
	}
	rawdata := strings.TrimPrefix(data, ced.Symbol)
	if ced.Symbol == XASSETS {
		rawdata = strings.TrimLeft(rawdata, "(Move(")
	} else if ced.Symbol == XFEEMANAGE {
		rawdata = strings.TrimLeft(rawdata, "(FeeForProducer(")
	}

	rawdata = strings.TrimRight(rawdata, "))")
	// 解析token
	start := strings.Index(rawdata, "[")
	end := strings.Index(rawdata, "]")
	if start >= 0 && end > 0 {
		tokenStr := rawdata[start+1 : end]
		tokenStr = strings.ReplaceAll(tokenStr, " ", "")
		tmpTokenStr := strings.Split(tokenStr, ",")
		var tokenBytes []byte
		for _, t := range tmpTokenStr {
			tt, err := decimal.NewFromString(t)
			if err != nil {
				return nil, err
			}
			tipt := byte(tt.IntPart())
			tokenBytes = append(tokenBytes, tipt)
		}
		ced.Token = string(tokenBytes)
	}
	//解析amount
	tmpDatas := strings.Split(rawdata, ",")
	amountStr := tmpDatas[len(tmpDatas)-1:]
	amount := strings.ReplaceAll(amountStr[0], " ", "")
	a, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, fmt.Errorf("parse amount error,err=%v", err)
	}
	ced.Amount = a.IntPart()

	return ced, nil
}
