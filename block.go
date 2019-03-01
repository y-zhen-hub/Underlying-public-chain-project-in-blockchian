package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"
)

type Block struct {
	Version int32
	PrevBlockHash []byte
	Merkleroot []byte
	Hash []byte
	Time int32
	Bits int32
	Nonce int32
	Transations []*Transation
	Height int32
}
var (
	maxnonce int32 = math.MaxInt32
)
//计算难度值函数
func CalculateTargetFast(bits []byte)  []byte{
	var result []byte

	//第一个字节，计算指数
	exponet:=bits[:1]
	fmt.Printf("%x\n",exponet)
	//计算后面3个系数
	coeffient:=bits[1:]
	fmt.Printf("%x\n",coeffient)

	str:=hex.EncodeToString(exponet)  //转换成字符串--18
	fmt.Printf("%s\n",str)
	//三个参数，第一个是要进行转换的字符串，第二个是字符串当前的进制数(16进制)，最后一个是要进行转换的进制数(十进制)，
	exp,_:=strconv.ParseInt(str,16,8)  //返回int8的整数
	fmt.Printf("%d\n",exp)

	result=append(bytes.Repeat([]byte{0x00},32-int(exp)),coeffient...)
	result=append(result,bytes.Repeat([]byte{0x00},32-len(result))...)
	return result
}



//区块序列化
func (block *Block) serialize()[]byte{  //作为结构体的函数
	result:=bytes.Join(  //字节拼接
		[][]byte{
			IntToHex(block.Version),
			block.PrevBlockHash,
			block.Merkleroot,
			IntToHex(block.Time),
			IntToHex(block.Bits),
			IntToHex(block.Nonce)},
		[]byte{},
	)
	return result  //序列化结果
}

//根据交易构建默克尔树--默克尔根
func (b*Block) createMerklTreeRoot(transation []*Transation)  {
	var tranHash [][]byte
	for _,tx:=range transation{  //遍历每个交易，存储在二维数组中
		tranHash=append(tranHash,tx.Hash())
	}
	mTree:=NewMerkleTree(tranHash)  //返回默克尔节点
	b.Merkleroot=mTree.RootNode.Data  //获得默克尔节点
}



func (b *Block) String()  {
	fmt.Printf("version:%s\n",strconv.FormatInt(int64(b.Version),10))
	fmt.Printf("Prev.BlockHash:%x\n", b.PrevBlockHash)
	fmt.Printf("Prev.Merkleroot:%x\n",b.Merkleroot)
	fmt.Printf("Prev.Hash:%x\n",b.Hash)
	fmt.Printf("Time:%s\n",strconv.FormatInt(int64(b.Time),10))
	fmt.Printf("Bits:%s\n",strconv.FormatInt(int64(b.Bits),10))
	fmt.Printf("Nonce:%s\n",strconv.FormatInt(int64(b.Nonce),10))
}

func (b *Block) Serialize() []byte{
	var encoded bytes.Buffer
	enc:=gob.NewEncoder(&encoded)
	err:=enc.Encode(b)
	if err!=nil{
		log.Panic(err)
	}
	return encoded.Bytes()
}

func (b *Block) Deserialize(d []byte) *Block {
	var block  Block
	decode:=gob.NewDecoder(bytes.NewReader(d))  //读操作
	err:=decode.Decode(&block)  //拿到结构体
	if err!=nil{
		log.Panic(err)
	}
	return &block
}

//创建并分会创始块
func NewGensisBlock(transaction []*Transation) *Block{
	block:=&Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		int32(time.Now().Unix()),
		404454260,
		0,
		transaction,
		0,
	}
	pow:=NewProofOfWork(block)
	nonce,hash:=pow.Run()
	block.Nonce=nonce
	block.Hash=hash
	block.String()
	return block

}

//与github上的不同
//func NewGenesisBlock(coinbase *Transation) *Block {
//	return NewBlock([]*Transation{coinbase}, []byte{})
//  return result.Bytes()
//}
func NewBlock(transanction []*Transation,prevBlockHash []byte,height int32) *Block {
	block:=&Block{
		2,
		prevBlockHash,
		[]byte{},
		[]byte{},
		int32(time.Now().Unix()),
		404454260,
		0,
		transanction,
		height,
	}
	pow:=NewProofOfWork(block)
	nonce,hash:=pow.Run()
	block.Hash=hash
	block.Nonce=nonce
	return block
}

