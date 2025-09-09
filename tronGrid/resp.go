package tronGrid

type RespTrc20HistoryItem struct {
	TransactionId string `json:"transaction_id"`
	TokenInfo     struct {
		Symbol   string `json:"symbol"`
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
		Name     string `json:"name"`
	} `json:"token_info"`
	BlockTimestamp int64  `json:"block_timestamp"`
	From           string `json:"from"`
	To             string `json:"to"`
	Type           string `json:"type"`
	Value          string `json:"value"`
}
type RespTrc20History struct {
	Data    []*RespTrc20HistoryItem `json:"data"`
	Success bool                    `json:"success"`
	Meta    struct {
		At       int64 `json:"at"`
		PageSize int   `json:"page_size"`
	} `json:"meta"`
}

type RespHistory struct {
	Data    []*Transaction `json:"data"`
	Success bool           `json:"success"`
	Meta    struct {
		At       int64 `json:"at"`
		PageSize int   `json:"page_size"`
	} `json:"meta"`
}

type Contract struct {
	Parameter struct {
		Value struct {
			OwnerAddress    string `json:"owner_address"`    // all
			Amount          int64  `json:"amount"`           // trx
			ToAddress       string `json:"to_address"`       // trx
			Data            string `json:"data"`             // usdt
			ContractAddress string `json:"contract_address"` // usdt
		} `json:"value"`
	} `json:"parameter"`
	Type string `json:"type"`
}

type Transaction struct {
	Ret []struct {
		ContractRet string `json:"contractRet"` // "SUCCESS"
	} `json:"ret"`
	TxID    string `json:"txID"`
	RawData struct {
		Data       string     `json:"data"`
		Contract   []Contract `json:"contract"`
		Expiration int64      `json:"expiration"`
		Timestamp  int64      `json:"timestamp"`
	} `json:"raw_data"`
}

type Block struct {
	BlockID     string `json:"blockID"`
	BlockHeader struct {
		RawData struct {
			Number    int64 `json:"number"`
			Timestamp int64 `json:"timestamp"`
		} `json:"raw_data"`
	} `json:"block_header"`
	Transactions []*Transaction `json:"transactions"`
}
