package main

//const dbFile = "blockchain.db"
func main() {
	// os.IsNotExist 判断err错误是否报告了一个文件或者目录不存在
	//if val,err := os.Stat(dbFile);os.IsNotExist(err) {
	//	fmt.Println(err)
	//} else {
	//	fmt.Println(val)
	//}
	//fmt.Println(err)

	//target := big.NewInt(1) // 初始化目标整数
	//target.Lsh(target,uint(256 - 18))
	//fmt.Println(target)
	//
	//var hashInt big.Int
	//
	//hash := sha256.Sum256([]byte("1231321321321321"))
	//hashInt.SetBytes(hash[:])
	//fmt.Println(hash[:])
	//fmt.Println(hashInt.Cmp(target))
	//isValid := hashInt.Cmp(pow.target) == -1
//	var arr  []int
//LABEL:
//	for i:=0;i<100;i++ {
//		for j:=0;j<100;j++ {
//			fmt.Println(j)
//			arr = append(arr,j)
//			if j==4 {
//				// continue 只跳出内层的循环 break整个循环结束
//				continue LABEL
//			}
//		}
//	}
//	fmt.Println(len(arr))
//	str := fmt.Sprintf("挖矿奖励给%s","zhangsan")

	cli := CLI{}
	cli.Run()
}
