package main

func main()  {
	//TestcreateMerkleTreeRoot()
	//target:=big.NewInt(1)  //初始化
	//target.Lsh(target,uint(256-targetBits))  //左移一定位数
	//fmt.Printf("%x",target.Bytes())  //打印的时候需要转化为字节类型
	//TestPow()  //测试是否可以正常挖矿
	//TestNewSerialize()  //测试是否可以正常序列化
	//NewGensisBlock()//测试是否可以正常创建世区块
	//TestBoltDB()  //验证区块链是否可以自动添加内容，自动生成区块链

	//验证是否能成功创建区块链---验证是否可以正常创建钱包及读取钱包地址
	bc :=NewBlockchain("13waVMhErsKBGG1kxhp3GhW39c1mArd81E")
	cli :=CLI{bc}
	cli.Run()


	//wallet :=NewWallet()
	//fmt.Printf("私钥：%x\n",wallet.PrivateKey.D.Bytes())
	//fmt.Printf("公钥：%x\n",wallet.Publickey)
	//fmt.Printf("地址: %x\n",wallet.GetAddress())
	//address,_:=hex.DecodeString("3159394165757a426d354651766f327950584665336d69506e356a4d6142355439")
	//fmt.Printf("地址相同？%t\n",ValidAddress(address))
}

