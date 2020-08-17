package model

type ChainXBlock struct {
	Block         Block  `json:"block"`
	Justification []byte `json:"justification"`
}

type Block struct {
	Extrinsics []string `json:"extrinsics"`
	Header     Header   `json:"header"`
}

type Header struct {
	ParentHash     string `json:"parentHash"`
	Number         string `json:"number"`
	StateRoot      string `json:"stateRoot"`
	ExtrinsicsRoot string `json:"extrinsicsRoot"`
	//Digest         interface{}   `json:"digest"`
}

type ChainXBlockResponse struct {
	Height     int64                      `json:"height"`
	ParentHash string                     `json:"parent_hash"`
	BlockHash  string                     `json:"block_hash"`
	Timestamp  int64                      `json:"timestamp"`
	Extrinsic  []*ChainXExtrinsicResponse `json:"extrinsic"`
}

type ChainXExtrinsicResponse struct {
	Type           string `json:"type"`   //Transfer or another
	Status         string `json:"status"` //success or fail
	Txid           string `json:"txid"`
	FromAddress    string `json:"from_address"`
	ToAddress      string `json:"to_address"`
	Amount         string `json:"amount"`
	Fee            string `json:"fee"`
	Signature      string `json:"signature"`
	Nonce          int64  `json:"nonce"`
	Era            string `json:"era"`
	ExtrinsicIndex int    `json:"extrinsic_index"`
	Token          string `json:"token"`
	Memo           string `json:"memo"`
}

type ChainXBlockEventResponse struct {
	BlockHash string              `json:"blockHash"`
	Events    map[string][]string `json:"events"`
	Number    int64               `json:"number"`
}
