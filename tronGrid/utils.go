package tronGrid

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"strings"

	"github.com/xnumb/tb/log"
)

var base58Alphabets = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func base58Encode(input []byte) []byte {
	x := big.NewInt(0).SetBytes(input)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := &big.Int{}
	var result []byte
	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, base58Alphabets[mod.Int64()])
	}
	reverseBytes(result)
	return result
}

func reverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
func fromHexAddress(hexAddress string) (string, error) {
	addrByte, err := hex.DecodeString(hexAddress)
	if err != nil {
		return "", err
	}

	sha := sha256.New()
	sha.Write(addrByte)
	shaStr := sha.Sum(nil)

	sha2 := sha256.New()
	sha2.Write(shaStr)
	shaStr2 := sha2.Sum(nil)

	addrByte = append(addrByte, shaStr2[:4]...)
	return string(base58Encode(addrByte)), nil
}
func decodeUsdtData(s string) (string, int64, error) {
	if len(s) < 136 {
		return "", 0, errors.New("解析错误, Data长度不符")
	}
	if len(s) > 136 {
		s = s[:136]
	}
	hexAddr := "41" + s[32:72] //strings.TrimLeft(s[:64], "0")
	hexAmount := strings.TrimLeft(s[72:], "0")
	addr, err := fromHexAddress(hexAddr)
	if err != nil {
		return "", 0, err
	}
	var amount int64 = 0
	if hexAmount != "" {
		amount, err = strconv.ParseInt(hexAmount, 16, 64)
		if err != nil {
			return "", 0, err
		}
	}
	return addr, amount, nil
}

func decodeTronData(data string) (string, error) {
	// 去除 0x 前缀
	hexStr := data
	if len(data) >= 2 && data[:2] == "0x" {
		hexStr = data[2:]
	}
	// 解码 hex 字符串
	decodedBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}

type Tx struct {
	Id     string
	Num    int64
	Ts     int64
	Token  Token
	Amount int64
	From   string
	To     string
	Remark string
}

type Txs []*Tx

func parseTx(rs []*Transaction, num int64, usdtContractHex string) Txs {
	var contract Contract
	var txs Txs
	for _, tx := range rs {
		if len(tx.RawData.Contract) > 0 {
			contract = tx.RawData.Contract[0]
		}
		ret := ""
		if len(tx.Ret) > 0 {
			ret = tx.Ret[0].ContractRet
		}
		if ret != "SUCCESS" {
			continue
		}
		if contract.Type != "TransferContract" && contract.Type != "TriggerSmartContract" {
			continue
		}
		r := Tx{
			Id:  tx.TxID,
			Num: num,
			Ts:  tx.RawData.Timestamp,
		}
		ctVal := contract.Parameter.Value
		r.From = ctVal.OwnerAddress
		if contract.Type == "TransferContract" { // trx交易
			toAddr, err := fromHexAddress(ctVal.ToAddress)
			if err != nil {
				continue
			}
			r.To = toAddr
			r.Amount = ctVal.Amount
			r.Token = TokenTrx
		} else if contract.Type == "TriggerSmartContract" { // usdt交易
			if ctVal.ContractAddress != usdtContractHex {
				continue
			}
			if len(ctVal.Data) <= 8 {
				continue
			}
			if ctVal.Data[:8] != "a9059cbb" {
				continue
			}
			addr, amount, err := decodeUsdtData(ctVal.Data)
			if err != nil {
				log.Err(err, "txid: %s, data: %s", tx.TxID, ctVal.Data)
				continue
			}
			r.To = addr
			r.Amount = amount
			r.Token = TokenUsdt
		} else {
			continue
		}
		remark, _ := decodeTronData(tx.RawData.Data)
		r.Remark = remark
		// r.from转成base58
		fromAddr, err := fromHexAddress(r.From)
		if err != nil {
			continue
		}
		r.From = fromAddr
		txs = append(txs, &r)
	}
	return txs
}
