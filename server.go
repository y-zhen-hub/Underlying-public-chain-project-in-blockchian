package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

var knownNodes =[]string{"localhost:3000"}
const nodeversion=0x00
const commandLength=12
var nodeAddress string
var blockInTransit [][]byte   //存储所有的区块信息

type Version struct {
	Version int
	BestHeight int32
	AddFrom string
}

type getblocks struct {
	Addrfrom string
}

type inv struct {
	AddrFrom string
	Type string
	Items [][]byte
}

type getdata struct {
	AddrFrom string
	Type string
	ID []byte
}

type blocksend struct {
	AddrFrom string
	Block []byte
}
func (ver *Version) String()  {
	fmt.Printf("version: %d\n",ver.Version)
	fmt.Printf("bestheight:%d\n",ver.BestHeight)
	fmt.Printf("addfrom: %s\n",ver.AddFrom)
}
//开启服务器
func StartServer(nodeID,minerAddress string,bc *BlockChain)  { //指定矿工ID，矿工地址
	nodeAddress=fmt.Sprintf("localhost:%s",nodeID)  //本地地址--将地址转换成字符串
	ln,err:=net.Listen("tcp",nodeAddress) //监听节点
	defer ln.Close()//监听完毕后关闭
	//fmt.Println("a1")
	//bc:=NewBlockchain("13waVMhErsKBGG1kxhp3GhW39c1mArd81E")
	if nodeAddress!=knownNodes[0]{
		sendVersion(knownNodes[0],bc) //若本地节点不在已知节点中，则发送版本信息
	}
	//fmt.Println("a2")
	for { //不断监听—死循环
		//fmt.Println("a3")
		conn,err2:=ln.Accept()//不断接收监听获得消息，
		if err2!=nil{
			log.Panic(err)
		}
		go handleConnection(conn,bc) //处理链接—添加协程
		//fmt.Println("a4")
	}
}

 //根据命令不同进行不同处理
func handleConnection(conn net.Conn, bc *BlockChain) {
	request,err:=ioutil.ReadAll(conn) //读取数据
	//fmt.Println("b1")
	if err!=nil{
		log.Panic(err)
	}

	command:=bytesToCommand(request[:commandLength]) //获得命令
	switch command {
	case "version":
		handleVersion(request,bc)
	case "getblocks":
		handleGetBlock(request,bc)
	case "inv":
		handleInv(request,bc)
	case "getdata":
		handleGetData(request,bc)
	case "block":
		handleBlock(request,bc)
	}
}

func handleBlock(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload blocksend

	buff.Write(request[commandLength:])
	dec:=gob.NewDecoder(&buff)
	err:=dec.Decode(&payload)
	if err!=nil{
		log.Panic(err)
	}

	blockdata:=payload.Block  //获得区块序列化数据
	block:=Deserialize(blockdata) //对数据进行反序列化
	bc.AddBlock(block) //若外部节点有新的区块数据，则本地数据库添加新的区块
	fmt.Printf("Receive a new Block")

	if len(blockInTransit)>0{ //若存储的数据信息不为空
		blockHash:=blockInTransit[0] //获得存储的最新哈希
		sendGetData(payload.AddrFrom,"block",blockHash) //将最新哈希发送给对方
		blockInTransit=blockInTransit[1:]  //剔除刚刚获得的最新区块哈希
	}else {
		set:=UTXOSet{bc}  //更新UTXO
		set.Reindex()  //将所有的未花费输出存储到数据库中
	}

}

func handleGetData(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getdata
	buff.Write(request[commandLength:])
	bec:=gob.NewDecoder(&buff)
	err:=bec.Decode(&payload)
	if err!=nil{
		log.Panic(err)
	}
	if payload.Type=="block"{
		block,err:=bc.GetBlock([]byte(payload.ID))
		if err!=nil{
			log.Panic(err)
		}
		sendBlock(payload.AddrFrom,&block)
	}
	
}

func sendBlock(addr string, block *Block) {
	payload:=gobEncode(blocksend{nodeAddress,block.Serialize()})
	request:=append(commandToBytes("block"),payload...)
	sendData(addr,request)
}

func handleInv(request []byte, bc *BlockChain) {
	var buff bytes.Buffer

	var payload inv
	buff.Write(request[commandLength:])

	dec:=gob.NewDecoder(&buff)
	err:=dec.Decode(&payload)

	if err!=nil{
		log.Panic(err)
	}
	fmt.Printf("Recieve inventory %d, %s\n", len(payload.Items),payload.Type)
	//for _,b:=range payload.Items{
	//	fmt.Printf("%x\n",b)
	//}

	if payload.Type=="block"{
		blockInTransit=payload.Items //存储所有的哈希值
		blockHash:=payload.Items[0] //外部节点的最后哈希，是最后一个区块的哈希值

		sendGetData(payload.AddrFrom,"block",blockHash)


		newInTransit:=[][]byte{}
		for _,b:=range blockInTransit{

			if bytes.Compare(b,blockHash)!=0{

				newInTransit=append(newInTransit,b) //剔除最高的区块哈希值

			}
		}
		blockInTransit=newInTransit

	}
}

func sendGetData(addr string, kind string, id []byte) {
	payload:=gobEncode(getdata{nodeAddress,kind,id}) //对结构体进行封装
	request:=append(commandToBytes("getdata"),payload...)
	sendData(addr,request)
}

func handleGetBlock(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getblocks
	buff.Write(request[commandLength:])
	dec:=gob.NewDecoder(&buff)
	err:=dec.Decode(&payload)
	if err!=nil{
		log.Panic(err)
	}
	block:=bc.Getblockhash()  //获得区块链中所有的哈希值
	sendInv(payload.Addrfrom,"block",block) //传递给另外一个节点区块链所有的哈希值
}

//封装传递的内容
func sendInv(addr string, kind string, items [][]byte) {
	inventory:=inv{nodeAddress,kind,items} //对结构体进行封装
	payload:=gobEncode(inventory)
	request:=append(commandToBytes("inv"),payload...)
	sendData(addr,request)

}

//版本不同时，更新版本
func handleVersion(request []byte, bc *BlockChain) {
	var buff bytes.Buffer  //定义缓存
	var payload Version  //定义版本
	buff.Write(request[commandLength:])  //在缓存中写入请求后面的内容
	dec:=gob.NewDecoder(&buff) //对请求内容进行解码
	err:=dec.Decode(&payload)  //解码
	if err!=nil{
		log.Panic(err)
	}
	myBestHeight:=bc.GetBestHeight()  //获得本地区块链的高度
	foreignerBestHeight:=payload.BestHeight //获得外部区块链高度
	payload.String()
	if myBestHeight<foreignerBestHeight{
		sendGetBlock(payload.AddFrom) //获得最新的区块链信息
	}else {
		sendVersion(payload.AddFrom,bc) //发送最新的区块链版本信息
	}
	if !nodeIsKnow(payload.AddFrom){  //将版本信息发送者的地址存储到已知节点中
		knownNodes=append(knownNodes,payload.AddFrom)
	}
}

//获得区块链信息函数
func sendGetBlock(address string) {
	payload:=gobEncode(getblocks{nodeAddress})
	request:=append(commandToBytes("getblocks"),payload...)
	sendData(address,request)
}

//判断是否在已知节点中
func nodeIsKnow(addr string) bool {
	for _,node:=range knownNodes{
		if node==addr{
			return true
		}
	}
	return false
}

//发送版本信息
func sendVersion(addr string, bc *BlockChain) {
	bestHeight:=bc.GetBestHeight()  //获得区块链的高度
	//对版本信息进行编码
	// addr是发送给对方的地址(接受者)，version中存储的自己(发送者)的地址
	payload:=gobEncode(Version{nodeversion,bestHeight,nodeAddress})
	//将版本命令添加到请求列表中
	request:=append(commandToBytes("version"),payload...) //version是一个命令
	sendData(addr,request) //向其他节点发送请求
}

//发送数据信息
func sendData(addr string, data []byte) { //将要发送的地址和数据
	con,err:=net.Dial("tcp",addr)  //和其他节点进行网络连接
	if err!=nil{
		fmt.Printf("%s is no available\n",addr)

		var updateNodes []string
		for _,node:=range knownNodes{  //遍历所有的已知节点
			if node!=addr{  //若不是将发送的节点，则将其添加到更新节点列表中
				updateNodes=append(updateNodes,node)
			}
		}
		knownNodes=updateNodes  //更新已知节点列表
	}
	defer con.Close() //延迟关闭--go语言特有的功能
	_,err=io.Copy(con,bytes.NewReader(data)) //传递数据
	if err!=nil{
		log.Panic(err)
	}
}

//将命令转换为字节
func commandToBytes(command string) []byte {
	var bytes [commandLength]byte //新建变量—字节长度
	for i,c:=range command{
		bytes[i]=byte(c)
	}
	return bytes[:]
}

//将字节转化为命令字符串
func bytesToCommand(bytes []byte) string {
	var command []byte
	for _,b:=range bytes{
		if b!=0x00{
			command=append(command,b)
		}
	}
	return fmt.Sprintf("%s",command)  //将命令转换成字符串
}

//序列化
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc:=gob.NewEncoder(&buff)
	err:=enc.Encode(data)
	if err!=nil{
		log.Panic(err)
	}
	return buff.Bytes()
}