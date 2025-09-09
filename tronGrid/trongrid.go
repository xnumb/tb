package tronGrid

import (
	"errors"
	"time"

	"github.com/xnumb/tb/fetch"
	"github.com/xnumb/tb/to"
)

const (
	Api             = "https://api.trongrid.io/"
	UsdtContract    = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	UsdtContractHex = "41a614f803b6fd780986a42c78ec9c7f77e6ded13c"
)
const (
	ShastaApi             = "https://api.shasta.trongrid.io/"
	ShastaUsdtContract    = "TG3XXyExBkPp9nzdajDZsozEu4BkaSJozs"
	ShastaUsdtContractHex = "4142a1e39aefa49290f2b3f9ed688d7cecf86cd6e0"
)

type Token int8

const (
	TokenTrx  Token = 0
	TokenUsdt Token = 1
)

func (e Token) Str() string {
	switch e {
	case TokenUsdt:
		return "USDT"
	case TokenTrx:
		return "TRX"
	default:
		return "未定义"
	}
}

type Client struct {
	debug           bool
	api             string
	key             string
	header          fetch.Header
	usdtContract    string
	usdtContractHex string
}

func New(key string, debug bool) *Client {
	api := Api
	usdtContract := UsdtContract
	usdtContractHex := UsdtContractHex
	if debug {
		api = ShastaApi
		usdtContract = ShastaUsdtContract
		usdtContractHex = ShastaUsdtContractHex
	}
	return &Client{
		debug: debug,
		api:   api,
		key:   key,
		header: fetch.Header{
			"TRON-PRO-API-KEY": key,
		},
		usdtContract:    usdtContract,
		usdtContractHex: usdtContractHex,
	}
}

// GetUsdtHistory direct> 0:双向 1:转入 2:转出
func (c *Client) GetUsdtHistory(expMin int64, count int, addr string, direct int) ([]*RespTrc20HistoryItem, error) {
	ts := to.S(time.Now().Unix()-expMin*60) + "000"
	directParam := ""
	if direct == 1 {
		directParam = "&only_to=true"
	} else if direct == 2 {
		directParam = "&only_from=true"
	}
	path := "v1/accounts/" + addr + "/transactions/trc20?limit=" + to.S(count) + "&only_confirmed=true&min_timestamp=" + ts + "&contract_address=" + c.usdtContract + directParam
	body, err := fetch.Get(c.api+path, c.header)
	if err != nil {
		return nil, err
	}
	r := &RespTrc20History{}
	if err = body.Parse(&r); err != nil {
		return nil, err
	}
	if !r.Success {
		return nil, errors.New("获取钱包交易历史发生错误")
	}
	var data []*RespTrc20HistoryItem
	// 防止欺骗
	for _, d := range r.Data {
		if d.Type == "Transfer" {
			data = append(data, d)
		}
	}
	return data, nil
}

func (c *Client) GetNowBlock() (int64, Txs, error) {
	res := &Block{}
	body, err := fetch.Get(c.api+"wallet/getnowblock", c.header)
	if err != nil {
		return 0, nil, err
	}
	if err = body.Parse(&res); err != nil {
		return 0, nil, err
	}
	txs := parseTx(res.Transactions, res.BlockHeader.RawData.Number, c.usdtContractHex)
	return res.BlockHeader.RawData.Number, txs, nil
}

func (c *Client) CheckAddrFmt(addr string) (bool, error) {
	res := struct {
		Result  bool   `json:"result"`
		Message string `json:"message"`
	}{}
	body, err := fetch.Post(c.api+"wallet/validateaddress", fetch.Params{
		"address": addr,
		"visible": true,
	}, c.header)
	if err != nil {
		return false, err
	}
	if err = body.Parse(&res); err != nil {
		return false, err
	}
	if res.Result {
		return res.Message == "Base58check format", nil
	}
	return false, errors.New(res.Message)
}

func (c *Client) CheckAddrActive(addr string) (bool, error) {
	res := struct {
		Address string `json:"address"`
	}{}
	body, err := fetch.Post(c.api+"wallet/getaccount", fetch.Params{
		"address": addr,
		"visible": true,
	}, c.header)
	if err != nil {
		return false, err
	}
	if err = body.Parse(&res); err != nil {
		return false, err
	}
	return res.Address != "", nil
}
