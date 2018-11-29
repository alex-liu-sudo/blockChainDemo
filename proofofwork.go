package main

import (
	"math"
	"math/big"
	"bytes"
	"fmt"
	"crypto/sha256"
)

var (
	maxNonce = math.MaxInt64 // int64的最大值
)

// 定义对比的位数
const targetBits = 16

type ProofOfWork struct {
	 block *Block // 区块
	 target *big.Int // 存储计算哈希对比的特定整数
}

// 创建有一个工作量证明的挖矿对象
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1) // 初始化目标整数
	target.Lsh(target,uint(256 - targetBits))

	return &ProofOfWork{block,target}
}

// 准备挖矿数据
func (pow *ProofOfWork) prepareData(nonce int) []byte {
		data := bytes.Join([][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),  // 时间戳转化为16进制
			IntToHex(int64(targetBits)), // 对比位数转化为16进制
			IntToHex(int64(nonce)), // 保存工作量证明的nonce
		},[]byte{})

		return data
}


// 执行挖矿
func (pow *ProofOfWork) Run() (int,[]byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Println("解题开始：")

	for nonce < maxNonce {
		data := pow.prepareData(nonce) // 准备挖矿数据
		hash = sha256.Sum256(data)  // 计算出哈希
		fmt.Printf("\r%x",hash) // 打印显示哈希
		hashInt.SetBytes(hash[:]) // 获取要对比的数据

		// 如果挖出的数据比当前的数据小 成功
		if hashInt.Cmp(pow.target) == -1 {
			fmt.Println("\n解题结束，题目答案是：",nonce)
			fmt.Printf("本次hash：%x\n",hash)
			break
		} else {
			nonce++
		}
	}
	fmt.Println("\n")
	return nonce,hash[:]
}

// 验证nonce是否合法
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce) // 准备好的数据

	// 对数据进行加密
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
