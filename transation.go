package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const subsidy  = 100  //在比特币中，该值是用区块总数量/210000，创始块的奖励是50BTC，每挖出210000个块后，奖励减半

type Transation struct {
	ID []byte
	Vin []TXInput  //输入的结构体
	Vout[]TXOutput  //输出结构体
}

type TXInput struct {
	TXid[]byte//前一个交易的哈希值，输入引用前一个输出的ID--哈希
	VoutIndex int //当前输出在交易中的索引，由于一笔交易可能有多个输出，需要指明具体是哪个
	Signature []byte  //标识输入可以解锁输出    //Bob //提供了可解锁输出结构中脚本公钥字段数据，若签名是正确的，则输出可以被解锁，解锁的值就可以被用于新产生的输出
	Pubkey []byte  //公钥--用于解锁输入，新添加--
}

type TXOutput struct {
	Value int  //输出的金额
	PubKeyHash []byte  //地址，输出有多少钱打到某个账户上  //Bob  //可以理解为要花费的钱，即要解锁的部分

}

type TXOutputs struct {  //存储所有的输出
	Outputs []TXOutput
}
//易于输出的调试函数
func (tx *Transation) String()  string{
	var lines []string
	lines=append(lines,fmt.Sprintf("--- Transantion: %x",tx.ID))
	for i, input :=range tx.Vin{
		lines=append(lines,fmt.Sprintf("Input: %d",i))
		lines=append(lines,fmt.Sprintf("TXID: %x",input.TXid))
		lines=append(lines,fmt.Sprintf("Out: %d",input.VoutIndex))
		lines=append(lines,fmt.Sprintf("Signature: %x",input.Signature))
	}
	for i,output:=range tx.Vout{
		lines=append(lines,fmt.Sprintf("Output: %d",i))
		lines=append(lines,fmt.Sprintf("Value: %d",output.Value))
		lines=append(lines,fmt.Sprintf("Script: %x",output.PubKeyHash))
	}
	return strings.Join(lines,"\n")
}

//序列化函数
func (tx *Transation)  Serialise() []byte {
	var encode bytes.Buffer
	enc:=gob.NewEncoder(&encode)
	//序列化
	err:=enc.Encode(tx)
	if err!=nil{
		log.Panic(err)
	}
	return encode.Bytes()
}

//计算交易的哈希值
//序列化后进行哈希
func (tx *Transation) Hash() []byte  {
	txcopy:=*tx
	txcopy.ID=[]byte{}
	//进行哈希运算
	hash:=sha256.Sum256(txcopy.Serialise())
	return hash[:]
}

//根据金额和地址新建一个输出
//存放交易的输出数据, ---已检查
func NewTXOutput(value int,address string)  *TXOutput{
	txo:=&TXOutput{value,nil}
	//txo.PubKeyHash=[]byte(address)  //明确打给某人的地址
	txo.Lock([]byte(address))   //这里得address只是个字符地址，需要将其转换成公钥哈希，
	return txo
}

//第一笔coinbase交易  ---已检查
func NewCoinbaseTX(to,data string)  *Transation{
	txin:=TXInput{[]byte{},-1,nil,[]byte(data)}
	txout:=NewTXOutput(subsidy,to)
	tx:=Transation{nil,[]TXInput{txin},[]TXOutput{*txout}} //初始化当前的交易
	tx.ID=tx.Hash() //对交易内容进行哈希
	return &tx
}

//下面两个函数，目的：用于计算UTXO遍历所有的输入输出
//对TXOutput结构体进行绑定,计算UTXO数据  --已检查
func (out *TXOutput)CanBeUnlockedWith(pubkeyhash []byte) bool  {
	return bytes.Compare(out.PubKeyHash,pubkeyhash)==0
}

//对TXInput结构体进行绑定,公钥哈希可以解锁输入---已检查
func (in *TXInput)CanUnlockOutputWith(unlockdata []byte) bool {
	lockinghash:=HashPubkey(in.Pubkey)
	return bytes.Compare(lockinghash,unlockdata)==0
}

//判断是否为第一笔交易
func (tx *Transation)IsCoinBase() bool {
	return len(tx.Vin)==1&&len(tx.Vin[0].TXid)==0&&tx.Vin[0].VoutIndex==-1
}

//进行数据签名--已检查
func (tx *Transation) Sign(privkey ecdsa.PrivateKey, prevTXs map[string]Transation) {
	if tx.IsCoinBase(){ //如果是COINBASE交易则，无须进行数据签名
		return
	}

	//进行检查,判断前一笔交易输出的哈希ID是否为空
	for _,vin :=range tx.Vin{  //遍历交易中所有的输入
		if prevTXs[hex.EncodeToString(vin.TXid)].ID==nil{ //说明前一笔交易输出哈希的ID不存在
			log.Panic("当前交易不存在")
		}
	}
	txcopy:=tx.TrimmedCopy() //生成交易副本
	for inID,vin:=range txcopy.Vin{
		prevTX:=prevTXs[hex.EncodeToString(vin.TXid)] //根据输入交易ID存储前一笔交易的结构体

		txcopy.Vin[inID].Signature=nil
		txcopy.Vin[inID].Pubkey=prevTX.Vout[vin.VoutIndex].PubKeyHash //这笔交易的这笔输入的引用的前一笔交易的输出的公钥哈希
		txcopy.ID=txcopy.Hash() //对交易内容进行哈希
		r,s,err:=ecdsa.Sign(rand.Reader,&privkey,txcopy.ID) //进行数据签名—API函数，三个部分，r, s, err
		if err!=err{
			log.Panic(err)
		}
		signature :=append(r.Bytes(),s.Bytes()...)
		tx.Vin[inID].Signature=signature  //将每笔输入进行数据签名
	}
}

//复制交易结构体--已检查
func (tx *Transation) TrimmedCopy() Transation {
	var inputs []TXInput
	var outputs []TXOutput

	for _,vin:=range tx.Vin{  //存储所有的输入
		inputs=append(inputs,TXInput{vin.TXid,vin.VoutIndex,nil,nil})
	}

	for _,vout :=range tx.Vout{ //存储所有的输出
		outputs=append(outputs,TXOutput{vout.Value,vout.PubKeyHash})
	}
	//存储交易结构体
	txCopy:=Transation{tx.ID,inputs,outputs}
	return txCopy
}

//验证交易是否引用了正确的输出
func (tx *Transation) Verify(prevTXs map[string]Transation) bool {
	if tx.IsCoinBase(){
		return true
	}
	for _,vin:=range tx.Vin{
		if prevTXs[hex.EncodeToString(vin.TXid)].ID==nil{
			log.Panic("当前区块中不存在该笔输入")
		}
	}
	txcopy:=tx.TrimmedCopy()
	//椭圆曲线
	curve:=elliptic.P256()
	for inID,vin:=range tx.Vin{
		prevTX:=prevTXs[hex.EncodeToString(vin.TXid)] //获得所有的输入
		txcopy.Vin[inID].Signature=nil//将所有输入的数据签名都设置为空
		txcopy.Vin[inID].Pubkey=prevTX.Vout[vin.VoutIndex].PubKeyHash //当前交易中引用的前一个交易输出对应的哈希
		txcopy.ID=txcopy.Hash() //对交易进行哈希运算

		r:=big.Int{}
		s:=big.Int{}
		siglen:=len(vin.Signature) //获得真实交易签名的长度
		r.SetBytes(vin.Signature[:(siglen/2)]) //将数字签名一分为二
		s.SetBytes(vin.Signature[(siglen/2):])

		x:=big.Int{}
		y:=big.Int{}
		keylen:=len(vin.Pubkey)
		x.SetBytes(vin.Pubkey[:(keylen/2)])
		y.SetBytes(vin.Pubkey[(keylen/2):])

		rawPubkey:=ecdsa.PublicKey{curve,&x,&y} //新建公钥结构体
		if ecdsa.Verify(&rawPubkey,txcopy.ID,&r,&s)==false{
			return false
		}
		txcopy.Vin[inID].Pubkey=nil
	}
	return true
}

//创建新的交易--未花费交易---已检查
func NewUTXOTransation(from,to string, amount int, bc *BlockChain) *Transation {
	var inputs []TXInput
	var outputs []TXOutput

	//通过钱包获得公钥哈希得第二种方法
	wallets,err:=NewWallets()
	if err!=nil{
		log.Panic(err)
	}
	wallet:=wallets.GetWallets(from)
	acc,validoutputs :=bc.FindSpendableOutputs(HashPubkey(wallet.Publickey),amount) //acc是指未花费的金额，后面参数是交易哈希与金额
	if acc<amount{ //若交易未花费的值小于转帐金额
		log.Panic("Error: Not enough funds")
	}
	for txid,outs :=range validoutputs{
		txID,err:=hex.DecodeString(txid)
		if err!=nil{
			log.Panic(err)
		}
		for _,out :=range outs{ //遍历所有的金额
			input :=TXInput{txID,out,nil,wallet.Publickey} //传入要输出的具体信息，
			inputs = append(inputs,input)  //将输出的具体信息存储到已输入列表中，
		}
	}
	outputs = append(outputs,*NewTXOutput(amount,to)) //将要转入to的具体信息存储到已输出列表
	if acc>amount{ //若可以花费的输出大于要转出的金额
		outputs = append(outputs,*NewTXOutput(acc-amount,from)) //将还未花费完的存储到已输出列表中
	}
	tx:=Transation{nil,inputs,outputs} //存储交易
	tx.ID=tx.Hash()

	//用私钥将交易进行数据签名
	bc.SignTransation(&tx,wallet.PrivateKey)
	return &tx
}

//通过地址获得公钥哈希  --已检查完
func (out *TXOutput) Lock (address []byte){
	decodeAddress :=Base58Decode(address) //反序列化地址
	pubkeyHash :=decodeAddress[1:len(decodeAddress)-4] //获得公钥哈希
	out.PubKeyHash=pubkeyHash //将公钥哈希传递给输出的公钥哈希
}

//对输出切片进行序列化
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer //定义一个缓冲区
	enc:=gob.NewEncoder(&buff)//添加缓冲区
	err:=enc.Encode(outs) //进行序列化
	if err!=nil{
		log.Panic(err)
	}
	return buff.Bytes()
}

//对输出切片进行反序列化
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs
	dec:=gob.NewDecoder(bytes.NewReader(data)) //对数据进行反序列化
	err:=dec.Decode(&outputs) //存储到输出的切片中
	if err!=nil{
		log.Panic(err)
	}
	return outputs
}