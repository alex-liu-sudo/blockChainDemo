package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
	"crypto/sha256"
)

// 定义一个区块
type Block struct {
	Id int64 // 区块Id
	Timestamp int64 // 时间戳
	Transactions []*Transaction // 交易的集合
	PrevBlockHash []byte // 上一块区块的哈希
	Hash []byte // 本区块的哈希
	Nonce int // 工作量证明
}


// 区块对象转化为二进制字符集 可以写到数据库中

func (block *Block) Serialize () []byte {
	var result bytes.Buffer // 定义存放字节集合的变量

	encoder := gob.NewEncoder(&result) // 创建二进制编码对象 可以使用该对象进行编码
	err := encoder.Encode(block) // 对区块进行编码
	if err != nil{
		log.Panic(err)
	}

	return result.Bytes() // 返回区块编码后的字节集
}

// 区块对象二进制字符集 解码 变回区块对象
func Deserialize(data []byte) *Block {
	var block Block // 解码后的区块对象
	decoder := gob.NewDecoder(bytes.NewReader(data)) // 创建一个二进制解码对象

	err := decoder.Decode(&block) // 解码二进制字节集为区块对象 block
	if err != nil {
		log.Panic(err)
	}
	return &block
}

// 创建创世区块
func NewGenesisBlock(coinbase *Transaction) *Block {
		// 创世区块
		block := &Block{0,time.Now().Unix(),[]*Transaction{coinbase},[]byte{},[]byte{},0}

		pow := NewProofOfWork(block) // 创世区块计算工作量证明
		nonce,hash := pow.Run()

		block.Hash = hash
		block.Nonce = nonce

		return block
}

// 创建区块
func NewBlock(transaction []*Transaction,prevBlock *Block) *Block {

	newBlockId := prevBlock.Id + 1

	// 定义一个区块
	block := &Block{newBlockId,time.Now().Unix(),transaction,prevBlock.Hash,[]byte{},0}

	// 挖矿附加这个区块
	pow := NewProofOfWork(block)

	nonce,hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return block
}

// 对于交易实现哈希计算
func (block *Block) HashTransactions() []byte {
		var txHashes [][]byte // 交易列表
		var txHash [32]byte // 交易哈希

		// 遍历区块的交易记录
		for _,tx := range block.Transactions {
				txHashes = append(txHashes,tx.ID)
		}
		txHash = sha256.Sum256(bytes.Join(txHashes,[]byte{}))

		return txHash[:]
}