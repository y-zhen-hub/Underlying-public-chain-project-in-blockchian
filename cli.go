package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct {
	bc *BlockChain
}


//验证参数的个数
func (cli *CLI) validateArgs (){
	if len(os.Args)<1 {
		fmt.Println("参数小于1")
		os.Exit(1)
	}
	fmt.Println(os.Args)
}

//添加区块
func (cli *CLI) addBlock()  {
	cli.bc.MineBlock([]*Transation{})
}

//当用户输入错误时，提醒用户
func (cli *CLI) printUsage()  {
	fmt.Println("USages")
	fmt.Println("addblock--增加区块")
	fmt.Println("printblock--打印区块链")
}

//打印区块练的信息
func (cli *CLI) printChain()  {
	cli.bc.printBlockchain()
}

func (cli *CLI) getBalance(address string)  {
	balance :=0
	//获得公钥哈希的第一种方式：通过地址获得公钥哈希
	decodeAddress :=Base58Decode([]byte(address))
	pubkeyhash := decodeAddress[1:len(decodeAddress)-4]


	//UTXOs :=cli.bc.FindUTXO(pubkeyhash)
	set:=UTXOSet{cli.bc} //这两步可以避免每次查询余额，耗费时间
	UTXOs:=set.FindUTXObyPubkeyHash(pubkeyhash)

	for _,out :=range UTXOs{
		balance+=out.Value
	}
	fmt.Printf("balance of '%s': %d\n",address,balance)
}

func (cli *CLI) send(from, to string, amount int)  {
	//newbc:=NewBlockchain(from)
	//defer newbc.db.Close()

	tx:=NewUTXOTransation(from,to,amount,cli.bc)
	cli.bc.MineBlock([] *Transation{tx})
	fmt.Printf("Success!")
}

func (cli *CLI) createWallet()  {
	wallets,_:=NewWallets()
	address:=wallets.CreateWallets()
	wallets.SaveToFile()
	fmt.Printf("your address: %s\n",address)
}

func (cli *CLI) listAddress() {
	wallets,err:=NewWallets()
	if err!=nil{
		log.Panic(err)
	}
	//fmt.Println("--->",wallets)  //测试代码
	addresses :=wallets.GetAddress()
	for _,address :=range addresses{
		fmt.Printf("the address is: %s\n",address)
	}
}

func (cli *CLI) getBestHeight() {
	fmt.Println(cli.bc.GetBestHeight())
}

func (cli *CLI) startNode(nodeID string, minnerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)

	if len(minnerAddress)>0{
		if ValidAddress([]byte(minnerAddress)){
			fmt.Println("minner is on ", minnerAddress)
		}else {
			log.Panic("error minner Address")
		}
	}
	fmt.Println("xxxx")
	StartServer(nodeID,minnerAddress,cli.bc)
}

//用户终端进行交互
func (cli *CLI) Run(){
	cli.validateArgs()

	nodeID:=os.Getenv("NODE_ID") //从系统中获得环境变量
	if nodeID==""{
		fmt.Printf("NODE_ID is not set\n")
		os.Exit(1)
	}

	addBlockCmd := flag.NewFlagSet("addblock",flag.ExitOnError)
	printBlockCmd := flag.NewFlagSet("printblock",flag.ExitOnError)

	getBalanceCmd :=flag.NewFlagSet("getbalance",flag.ExitOnError)
	getBalanceAddress:=getBalanceCmd.String("address","","the address to get balance of ")

	startNodeCmd :=flag.NewFlagSet("startnode",flag.ExitOnError)
	startNodeMinner :=startNodeCmd.String("minner","","minner address")

	sendCmd :=flag.NewFlagSet("send",flag.ExitOnError)
	sendFrom :=sendCmd.String("from","","source wallet address")
	sendTo :=sendCmd.String("to","","Destination wallet address")
	sendAmount :=sendCmd.Int("amount",0,"Amount to send")

	createWalletCmd :=flag.NewFlagSet("createwallet",flag.ExitOnError)
	listaddressCmd :=flag.NewFlagSet("listaddress",flag.ExitOnError)
	getbestheightCmd :=flag.NewFlagSet("getbestheight",flag.ExitOnError)

	switch os.Args[1] {
	case "startnode":
		err:=startNodeCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}
	case "getbestheight":
		err:=getbestheightCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}

	case "createwallet":
		err:=createWalletCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}

	case "listaddress":
		err:=listaddressCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}
	case "send":
		err:=sendCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}
	case "getbalance":
		err:=getBalanceCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}
	case "addblock":
		err:=addBlockCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}
	case "printblock":
		err:=printBlockCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}
		default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed(){
		cli.addBlock()
	}
	if printBlockCmd.Parsed(){
		cli.printChain()
	}
	if getBalanceCmd.Parsed(){
		if *getBalanceAddress==""{
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}
	if sendCmd.Parsed(){
		if *sendFrom=="" || *sendTo=="" || *sendAmount<=0 {
			os.Exit(1)
		}
		cli.send(*sendFrom,*sendTo,*sendAmount)
	}

	if createWalletCmd.Parsed(){
		cli.createWallet()
	}
	if listaddressCmd.Parsed(){
		cli.listAddress()
	}
	if getbestheightCmd.Parsed(){
		cli.getBestHeight()
	}
	if startNodeCmd.Parsed(){
		nodeID:=os.Getenv("NODE_ID")
		if nodeID=="" {
			startNodeCmd.Usage()  //打印提示用例
			os.Exit(1)
		}
		cli.startNode(nodeID,*startNodeMinner)
	}
}



