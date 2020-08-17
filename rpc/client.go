package rpc

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JFJun/chainX-go/model"
	"github.com/JFJun/chainX-go/tx"
	"github.com/JFJun/chainX-go/util"
	"golang.org/x/crypto/blake2b"
	"strconv"
	"strings"
)

type Client struct {
	Rpc         *util.RpcClient
	CoinType    string
	GenesisHash string
}

func New(url, user, password string) (*Client, error) {
	client := new(Client)
	if strings.HasPrefix(url, "wss") {
		//todo 连接websocket
		return client, errors.New("do not support websocket")
	}
	client.Rpc = util.New(url, user, password)

	//初始化运行版本

	return client, nil
}

func (client *Client) GetBlockByNumber(height int64) (*model.ChainXBlockResponse, error) {
	var (
		respData []byte
		err      error
	)
	respData, err = client.Rpc.SendRequest("chain_getBlockHash", []interface{}{height})
	if err != nil || len(respData) == 0 {
		return nil, fmt.Errorf("get block hash error,err=%v", err)
	}
	blockHash := string(respData)
	return client.GetBlockByHash(blockHash)
}

func (client *Client) GetBlockByHash(blockHash string) (*model.ChainXBlockResponse, error) {
	var (
		respData []byte
		err      error
	)
	respData, err = client.Rpc.SendRequest("chain_getBlock", []interface{}{blockHash})
	if err != nil || len(respData) == 0 {
		return nil, fmt.Errorf("get block error,err=%v", err)
	}
	var block model.ChainXBlock
	err = json.Unmarshal(respData, &block)
	if err != nil {
		return nil, fmt.Errorf("parse block error")
	}
	blockResp := new(model.ChainXBlockResponse)
	number, _ := strconv.ParseInt(util.RemoveHex0x(block.Block.Header.Number), 16, 64)
	blockResp.Height = number
	blockResp.ParentHash = block.Block.Header.ParentHash
	blockResp.BlockHash = blockHash
	if len(block.Block.Extrinsics) > 0 { //todo parse extrinsic
		err = client.parseBlockByExtrinsic(block.Block.Extrinsics, blockResp)
		if err != nil {
			return blockResp, fmt.Errorf("parse block extrinsic error,Err=[%v]", err)
		}
		//解析事件event
		err = client.parseTxEventByBlockHash(blockResp)
		if err != nil {
			return blockResp, fmt.Errorf("parse block event error,Err=%v", err)
		}
	}
	return blockResp, nil
}

func (client *Client) parseBlockByExtrinsic(extrinsics []string, blockResponse *model.ChainXBlockResponse) error {
	for i, extrinsic := range extrinsics {
		if extrinsic == "" {
			return errors.New("extrinsic is null")
		}
		data, err := hex.DecodeString(util.RemoveHex0x(extrinsic))
		if err != nil {
			return fmt.Errorf("hex decode extrinsic error,Err=%v", err)
		}
		ex := tx.NewChainXExtrinsic(data)
		err = ex.ParseChainXExtrinsic()
		if err != nil {
			return fmt.Errorf("parse extrinsic error,Err=%v", err)
		}
		if ex.CallIndex == tx.CallIdTimestamp {
			blockResponse.Timestamp = ex.Timestamp
		} else if ex.CallIndex == tx.CallIdTransfer {
			blockEx := new(model.ChainXExtrinsicResponse)
			blockEx.Type = "transfer"
			blockEx.FromAddress = ex.From
			blockEx.ToAddress = ex.To
			blockEx.Token = ex.Token
			blockEx.Memo = ex.Memo
			blockEx.Signature = ex.Signature
			blockEx.Nonce = int64(ex.Nonce)
			blockEx.Era = ex.Era
			blockEx.ExtrinsicIndex = i
			blockEx.Amount = fmt.Sprintf("%d", ex.Amount)
			blockEx.Txid = client.createTxHash(extrinsic)
			blockResponse.Extrinsic = append(blockResponse.Extrinsic, blockEx)
		}
	}
	return nil
}

func (client *Client) parseTxEventByBlockHash(blockResponse *model.ChainXBlockResponse) error {
	if blockResponse == nil {
		return errors.New("block is null ptr")
	}
	blockHash := blockResponse.BlockHash
	var (
		respData []byte
		err      error
	)
	respData, err = client.Rpc.SendRequest("chainx_getExtrinsicsEventsByBlockHash", []interface{}{blockHash})
	if err != nil || len(respData) == 0 {
		return fmt.Errorf("get blockhash=[%s] event error,err=%v", blockHash, err)
	}
	var eventResp model.ChainXBlockEventResponse
	err = json.Unmarshal(respData, &eventResp)
	if err != nil {
		return fmt.Errorf("blockHash=[%s] paese event error,Err=[%v]", blockHash, err)
	}
	if eventResp.Events == nil {
		return nil
	}
	var extrinsicArray []*model.ChainXExtrinsicResponse
	for k, event := range eventResp.Events {
		if len(event) <= 1 {
			continue
		}
		extrinsic := new(model.ChainXExtrinsicResponse)
		extrinsic.Status = "failed"
		extrinsicIdx, _ := strconv.ParseInt(k, 10, 32)
		extrinsic.ExtrinsicIndex = int(extrinsicIdx)
		var feeArrey []*model.ChainXEventData
		for _, e := range event {
			if e == "system(ExtrinsicSuccess)" {
				extrinsic.Status = "success"
			} else if strings.Contains(e, model.XFEEMANAGE) {
				//解析手续费
				ced, _ := model.ParseChainXEventData(e)
				if ced != nil {
					feeArrey = append(feeArrey, ced)
				}
			}
		}
		var fee int64
		if len(feeArrey) > 0 {
			for _, f := range feeArrey {
				fee += f.Amount
			}
		}
		extrinsic.Fee = fmt.Sprintf("%d", fee)
		extrinsicArray = append(extrinsicArray, extrinsic)
	}
	//和blockresponse进行比较
	for _, ex := range extrinsicArray {
		for _, extrinsic := range blockResponse.Extrinsic {
			if ex.ExtrinsicIndex == extrinsic.ExtrinsicIndex {
				extrinsic.Status = ex.Status
				extrinsic.Fee = ex.Fee
			}
		}
	}
	return nil
}
func (*Client) createTxHash(extrinsic string) string {
	data, _ := hex.DecodeString(util.RemoveHex0x(extrinsic))
	d := blake2b.Sum256(data)
	return "0x" + hex.EncodeToString(d[:])
}
