package main

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"encoding/hex"
	"bytes"
	"encoding/gob"
	"crypto/sha256"
)

// 交易类
// 区块链中 先有输入 后有输出

const subsidy = 10 // 奖励 旷工挖矿给予的奖励

// 定义输入类型 TXInput 收钱
type TXInput struct {
	Txid []byte // 交易的id
	Vout int // 被花费的utxo索引号
	ScriptSig string // 保存一个任意用户定义的钱包地址 一个满足使用该utxo所需要的脚本
}

// 判断是否可以锁定
func (input *TXInput) CanUnlockOutPutWith(unlockingData string) bool {
		return input.ScriptSig == unlockingData
}

// 检查交易事务是否为coinbase
func (tx *Transaction) IsCoinBase() bool {
	// 必须满足三个条件
	// 1 必须有交易的输入 2 不存在交易的id 3 交易的vout outputs索引为-1
	//	txin := TXInput{[]byte{},-1,data} // 输入奖励 这边定义的条件
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}



// 定义输出类型 TXOutput 给钱
type TXOutput struct {
	Value int // 保存了币 交易的金额
	ScriptPubkey string // 保存用户定义的钱包地址 一个定义了使用该输出所需条件的脚本
}

// 是否可以解锁输出
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
		return out.ScriptPubkey == unlockingData // 判断用户的公钥 是否是这个公钥
}

// 交易类 编号 输入 输出
type Transaction struct {
	ID []byte // 编号
	Vin []TXInput // 输入
	Vout []TXOutput // 输出
}


// 矿工挖矿挖出区块后奖励币 coinbase
func NewCoinBaseTX(to,data string) *Transaction {
		if data == "" {
			data = fmt.Sprintf("挖矿奖励给%s",to)
		}

		// 挖矿奖励 没有人给钱 所以 TXInput 中的Vout output索引为-1
		txin := TXInput{[]byte{},-1,data} // 输入奖励
		txout := TXOutput{subsidy,to} // value 给多少 to给谁

		// 挖矿产生的交易
		tx := Transaction{nil,[]TXInput{txin},[]TXOutput{txout}}

		// 返回本次交易
		return &tx
}

// 点对点 普通交易
func NewUTXOTransAction(from,to string,amount int,bc *BlockChain) *Transaction {
		var inputs []TXInput // 输入
		var outputs []TXOutput // 输出


		acc,validOutputs := bc.FindSpendableOutputs(from,amount)
		if acc < amount {
			log.Panic("您可用的金额 小于本次支出的金额")
		}

		// 循环遍历无效输出
		for txid,outs := range validOutputs {
			txID,err := hex.DecodeString(txid) // 解码
			if err != err {
				log.Panic("处理错误")
			}

			for _,out := range outs {
				input := TXInput{txID,out,from} // 输入交易
				inputs = append(inputs,input) // 输出交易
			}

			// 交易叠加

			// 本次交易的支出
			outputs = append(outputs,TXOutput{amount,to})
			if acc > amount {
				// 找零给自己
				outputs = append(outputs,TXOutput{acc - amount,from})
			}
		}
		tx := Transaction{nil,inputs,outputs} // 一次交易
		tx.SetID() // 设置交易id
		return &tx
}

func (tx *Transaction) SetID() {
		var encoded bytes.Buffer // 开辟内存 存储字节集
		var hash [32]byte // 开辟32位的byte字节集

		enc := gob.NewEncoder(&encoded) // 编码对象
		err := enc.Encode(tx) // 对本次交易进行编码
		if err != nil {
			log.Panic(err)
		}
		hash = sha256.Sum256(encoded.Bytes()) // 计算本次交易的哈希

		tx.ID = hash[:] // 设置交易id
}

