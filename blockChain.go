package main

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
	"os"
	"fmt"
)

const dbFile = "blockchain.db" // 数据库文件名 路径为当前路径
const blockBucket = "blocks" // 数据表名称
const genesisCoinbaseData = "创币交易 第一笔交易 创始一个币" // 创币交易

// 区块链定义
type BlockChain struct {
	tip []byte // 二进制数据
	db *bolt.DB // 数据库
}

// 迭代器定义
type BlockChainIterator struct {
	currentHash []byte // 当前的哈希地址
	db *bolt.DB
}

// 挖矿带来的交易
func (bc *BlockChain) MineBlock(transaction []*Transaction) {
		var lastHash []byte // 最后一个区块的哈希
		var lastBlock *Block // 最后一个区块

		err := bc.db.View(func (tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(blockBucket)) // 打开桶
			lastHash = bucket.Get([]byte("1")) // 最后一个区块的哈希
			lastBlock = Deserialize(bucket.Get(lastHash))

			return nil
		})
		if err != nil {
			log.Panic(err)
		}

		// 创建一个新的区块
		newBlock := NewBlock(transaction,lastBlock) // 创建一个新的区块

		err = bc.db.Update(func (tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(blockBucket))

			// 将新的区块 格式化 写入数据库 加入到区块链
			err := bucket.Put(newBlock.Hash,newBlock.Serialize())
			if err != nil {
				log.Panic(err) // 区块写入错误
			}

			// 将最新的一块的哈希写入数据库
			err = bucket.Put([]byte("1"),newBlock.Hash) // 压入 保存最后一个哈希
			if err != nil {
				log.Panic(err)
			}

			bc.tip = newBlock.Hash
			return nil
		})
		if err != nil {
			log.Panic(err)
		}
}


// 获取没有使用的输出
func  (bc *BlockChain) FindSpendableOutputs(address string,amount int) (int,map[string][]int) {

	unspentOutputs := make(map[string][]int) // 开辟空间 未耗尽的输出记录列表

	unspentTxs := bc.FindUnspentTransactions(address) // 根据地址查找所有未花费的交易

	accmulated := 0 // 累计 账户金额

Work:
	for _,tx := range unspentTxs {
			txID := hex.EncodeToString(tx.ID) // 获取交易编号
			for outindex,out := range tx.Vout { // 遍历交易中的所有输出
					if out.CanBeUnlockedWith(address) && accmulated < amount { // 判断是否可以解锁本次输出
							accmulated += out.Value // 统计金额
							unspentOutputs[txID] = append(unspentOutputs[txID],outindex) // 叠加序列
							if accmulated >= amount {
								break Work // break Work 跳出后 不再执行标签对应的for循环 跳出到Work标识
							}

					}

			}
	}
	return accmulated,unspentOutputs
}

// 查找没有使用的交易输出
func (bc *BlockChain) FindUTXO(address string) []TXOutput {

	// 存储所有未花费的交易输出
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address) // 查找没有使用的交易列表

	for _,tx := range unspentTransactions {
			for _,out := range tx.Vout { // 遍历交易中的交易输出
					if out.CanBeUnlockedWith(address) { // 判断本次输出是否锁定
							UTXOs = append(UTXOs,out) // 加入数据
					}
			}
	}

	return UTXOs

}



// 获取没使用输出的交易列表
func (bc *BlockChain) FindUnspentTransactions(address string) []Transaction {

		var unspentTXs []Transaction // 没有使用输出的交易列表

		spentTXOS := make(map[string][]int) // 开辟内存

		// 定义一个区块链的迭代器 迭代循环整个区块链
		bci := bc.Iterator()

		// 循环这个迭代器
		for {
			block := bci.next() // 循环到下一个

			// 遍历区块中所有的交易
			for _,tx := range block.Transactions {
				txID := hex.EncodeToString(tx.ID) // 获取本次交易的编号 转化为字符串
			OutPuts:
				// 循环区块中所有的Vout outputs
				for outindex,out := range tx.Vout { // outindex 输出索引

						// 判断本次输出是否已经被使用
						if spentTXOS[txID] != nil {

							// 循环已经使用了的UTXO
							for _,spentOut := range spentTXOS[txID] {
									// 如果循环到的输出的索引 在已经消费的UTXO中 跳至 OutPuts重新执行
									if spentOut == outindex {
										continue OutPuts
									}
							}
						}

						// 判断地址是否可以解锁输出
						if out.CanBeUnlockedWith(address) {
							unspentTXs = append(unspentTXs,*tx)  // 如果可以解锁的话 将本次交易加入未使用输出的交易列表
 						}
				}

				// 判断是否是挖矿奖励
				if tx.IsCoinBase() == false { // 如果不是挖矿交易
					for _,in := range tx.Vin { // 循环交易中的所有输入

						if in.CanUnlockOutPutWith(address) { // 判断是否解锁输出
									// vin 代表
									inTxID := hex.EncodeToString(in.Txid)
									// 将已经使用的vout加入 spentTXOS 已经使用的交易列表
									spentTXOS[inTxID] = append(spentTXOS[inTxID],in.Vout)
						}
					}

				}

			}

			if len(block.PrevBlockHash) == 0 { // 最后一块 创世区块的时候 跳出循环 结束循环
					break
			}

		}


		return unspentTXs

}

// 定义迭代器 迭代整个区块链
func (bc *BlockChain) Iterator() *BlockChainIterator {
	bcit := &BlockChainIterator{bc.tip,bc.db}
	return bcit
}

// 根据迭代器取得下一个区块  返回下一个区块
func (it *BlockChainIterator) next() *Block {
	var block  *Block
	// 使用数据库查询
	err := it.db.View(func (tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))

		encodedBlock := bucket.Get(it.currentHash) // 根据当前的索引获取当前的区块 encode后的区块内容

		block = Deserialize(encodedBlock) // decode 解码区块内容

		return nil

	})
	if err != nil {
		log.Panic(err)
	}

	// 从最后一个区块开始往前进行迭代 改变指针位置
	it.currentHash = block.PrevBlockHash

	return block
}


// 新建一个区块链 打开之前创建的区块链
func NewBlockChain(address string) *BlockChain {
		// 判断数据库是否存在
		if dbExists() == false {
			fmt.Println("数据库不存在 请创建一个数据库")
			os.Exit(1)
		}

		var tip []byte // 存储区块链的二进制数据

		// 打开数据库 bolt.Open p1 数据库路径 mode 0600 打开方式 options 选项
		db,err := bolt.Open(dbFile,0600,nil)
		if err != nil {
			log.Panic(err)
		}

		// 数据库更新
		err = db.Update(func (tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(blockBucket)) // 按照bucket 名称打开数据库的表格

			tip = bucket.Get([]byte("1")) // 取出最后一条数据

			return nil

		})

		if err != nil {
			log.Panic(err) // 处理数据库更新错误
		}

		bc := BlockChain{tip,db} // 创建一个区块链
		return &bc
}

// 创建一个区块链创建一个数据库
func CreateBlockChain(address string) *BlockChain {
	 if dbExists() {
	 		// 数据库已经存在 不需要创建
	 		fmt.Println("数据已经存在，不需要进行创建")
	 		os.Exit(1)
	 }

	 var tip []byte // tip 存储区块链的二进制数据

	 // 打开数据库
	 db,err := bolt.Open(dbFile,0600,nil)
	 if err != nil {
	 	log.Panic(err) // 处理数据库打开错误
	 }

	 // 更新数据库
	 err = db.Update(func (tx *bolt.Tx) error {

	 	cbtx := NewCoinBaseTX(address,genesisCoinbaseData) // 创世区块的事务交易
	 	// 创建创世区块的区块
	 	genesis := NewGenesisBlock(cbtx)

	 	// bolt 创建bucket
	 	bucket,err := tx.CreateBucket([]byte(blockBucket))
	 	if err != nil {
	 		log.Panic(err)
		}

	 	// 把数据存储到数据库 []byte("1") 最后一个块
	 	err = bucket.Put(genesis.Hash,genesis.Serialize()) // 把区块格式化后存入数据库中
	 	if err != nil {
	 		log.Panic(err) // 数据存入失败
		}
	 	err = bucket.Put([]byte("1"),genesis.Hash)
	 	if err != nil {
	 		log.Panic(err) // 数据哈希存入失败
		}

	 	tip = genesis.Hash

	 	return nil
	 })

	 bc := BlockChain{tip,db} // 创建一个区块链
	 return &bc
}

// 判断数据库是否存在
func dbExists() bool {
		// os.Stat 返回文件描述相关的信息
		if _,err := os.Stat(dbFile);os.IsNotExist(err) {
			return false
		}
		return true
}
