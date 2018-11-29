package main

import (
	"fmt"
	"os"
	"strconv"
	"flag"
	"log"
)

// CLI 命令行接口
type CLI struct {
		bc *BlockChain
}

// 创建区块链
func (cli *CLI) createBlockChain(address string) {
		bc := CreateBlockChain(address) // 根据地址创建区块链
		bc.db.Close() // 创建完毕 关闭数据库

		fmt.Println("创建区块链成功",address)
}

// 查看区块链
func (cli *CLI) showChain() {
	 bc := NewBlockChain("")
	 bci := bc.Iterator() // 迭代器
	fmt.Println("\n")
	 for {
	 	block := bci.next()
	 	fmt.Printf("上一块区块的哈希：%x \n",block.PrevBlockHash)

	 	fmt.Printf("当前区块的ID：%d \n",block.Id)
	 	fmt.Printf("当前区块的哈希：%x \n",block.Hash)
	 	fmt.Printf("当前区块的工作量证明：%d \n",block.Nonce)
	 	fmt.Println(*block.Transactions[0])

	 	// 工作量证明
	 	pow := NewProofOfWork(block)
	 	fmt.Printf("pow %s \n",strconv.FormatBool(pow.Validate()))
		if block.PrevBlockHash == nil {
			fmt.Println("\n----已经循环到创世区块----\n")
			os.Exit(1)
		}
	 	fmt.Println("\n")
	 }
}

// 获取余额
func (cli *CLI) getBalance(address string) {
	bc := NewBlockChain(address)
	UTXOs := bc.FindUTXO(address)
	balance := 0
	for _,out := range UTXOs {
			balance += out.Value
	}

	fmt.Printf("当前 %s 的余额为 %d",address,balance)

}

// 转账
func (cli *CLI) send(from,to string,amount int) {
	bc := NewBlockChain(from)
	defer bc.db.Close()

	// 创建一笔交易
	tx := NewUTXOTransAction(from,to,amount,bc)

	// 挖矿 将交易打包进区块
	bc.MineBlock([]*Transaction{tx})

	fmt.Println("交易成功")
}

// cli 用法
func (cli *CLI) printUsage() {
		fmt.Println("用法如下")
		fmt.Println("send -from From -to To -amount Amount 转账金额")
		fmt.Println("getbalance -address 根据地址查询金额")
		fmt.Println("createblockchain -address 根据输入的地址创建一个区块链 ")
		fmt.Println("showchain 查看区块链")
}


func (cli *CLI) validateAtgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1) // os.Exit(1) 结束代码执行 返回0 正常  其余都是异常
	}
}

// cli执行
func (cli *CLI) Run() {
	cli.validateAtgs() // 效验命令

	// 处理命令行参数
	createblockchaincmd := flag.NewFlagSet("createblockchain",flag.ExitOnError)
	showchaincmd := flag.NewFlagSet("showchain",flag.ExitOnError)
	sendcmd := flag.NewFlagSet("send",flag.ExitOnError)
	getbalancecmd := flag.NewFlagSet("getbalance",flag.ExitOnError)

	getbalanceaddress := getbalancecmd.String("address","","查询地址")
	createblockchainaddress := createblockchaincmd.String("address","","查询地址")
	sendform := sendcmd.String("from","","付钱的")
	sendto := sendcmd.String("to","","收钱的")
	sendamount := sendcmd.Int("amount",0,"金额")

	// 解析命令
	switch os.Args[1] {
	case "getbalance":
		err := getbalancecmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createblockchaincmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "showchain":
		err := showchaincmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendcmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	// 执行
	if createblockchaincmd.Parsed() {
		if *createblockchainaddress == "" {
			createblockchaincmd.Usage()
			os.Exit(1)
		} else {
			cli.createBlockChain(*createblockchainaddress) // 创建区块链
		}
	}

	// 显示区块链
	if showchaincmd.Parsed() {
		cli.showChain() // 显示区块链
	}

	// 转账
	if sendcmd.Parsed() {
			if *sendform == "" || *sendto == "" || *sendamount == 0 {
					sendcmd.Usage()
					os.Exit(1)
			}

			cli.send(*sendform,*sendto,*sendamount)
	}

	// 创建区块链
	if createblockchaincmd.Parsed() {
		if *createblockchainaddress == "" {
			createblockchaincmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createblockchainaddress)
	}

	// 获取未花费的所有的UTXO总额
	if getbalancecmd.Parsed() {
			if *getbalanceaddress == "" {
					getbalancecmd.Usage()
					os.Exit(1)
			}
			cli.getBalance(*getbalanceaddress)
	}


}