package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile  ="wallet.dat" //用于存储钱包内容--私钥

type Wallets struct {  //将钱包和地址绑定在一块
	Walletsstore map[string] *Wallet
}

//新添加的钱包
func NewWallets()(*Wallets,error)  {
	wallets := &Wallets{}
	wallets.Walletsstore=make(map[string]*Wallet)
	err:=wallets.LoadFromFile()
	return wallets,err
}

//创建钱包
func (ws *Wallets) CreateWallets() string {
	wallet:=NewWallet()  //获得新的钱包
	address :=fmt.Sprintf("%s",wallet.GetAddress()) //将地址转换成字符存
	ws.Walletsstore[address]=wallet //完成映射存储
	//ws.SaveToFile()  //测试代码
	return address
}

//获得钱包地址对应的具体内容
func (ws *Wallets)GetWallets(address string) Wallet {
	return *ws.Walletsstore[address]
}

//获得所有的钱包地址
func (ws *Wallets) GetAddress() []string {
	var addresses []string  //用于存储所有的钱包地址
	for address:=range ws.Walletsstore{ //遍历所有的地址
		addresses =append(addresses,address) //进行地址存储
	}
	return addresses
}

//存储文件
func (ws *Wallets) SaveToFile()  {
	var content bytes.Buffer //创建一个缓存区
	gob.Register(elliptic.P256()) //提示接口用的椭圆曲线算法
	//将钱包结构体序列化
	encoder:=gob.NewEncoder(&content)  //告诉程序将要写入content中
	err:=encoder.Encode(ws) //将括号中的内容写入encoder中
	if err!=nil{
		log.Panic(err)
	}
	err = ioutil.WriteFile(walletFile,content.Bytes(),0777)  //将内容写入文件中，第三个参数的权限—0777—最高的权限，任何人都可以读取、修改
	if err!=nil{
		log.Panic(err)
	}
}

//读取文件
func (ws *Wallets) LoadFromFile() error {
	if _,err :=os.Stat(walletFile);os.IsNotExist(err){ //判断是否存在该文件，若存在则继续运行，若不存在，则直接报错
		return err
	}
	fileContent,err:=ioutil.ReadFile(walletFile)  //由于接收读取到的文件
	if err!=nil{
		log.Panic(err)
	}
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder :=gob.NewDecoder(bytes.NewReader(fileContent)) //反序列化
	err =decoder.Decode(&wallets)
	if err!=nil{
		log.Panic(err)
	}
	ws.Walletsstore=wallets.Walletsstore
	//fmt.Println(ws.Walletsstore)
	return nil	
}