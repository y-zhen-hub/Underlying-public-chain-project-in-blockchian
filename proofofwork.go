package main

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

type ProofOfWork struct {
	block *Block
	target *big.Int
}

const targetBits  = 16

//接收区块数据，为后面挖矿做准备
func NewProofOfWork(b *Block) *ProofOfWork {
	target:=big.NewInt(1)  //初始化
	target.Lsh(target,uint(256-targetBits))  //左移一定位数
	//fmt.Printf("%x",target.Bytes())  //打印的时候需要转化为字节类型
	pow:=&ProofOfWork{b,target}//初始化

	return pow
}

//类似于Serialise函数，序列化区块数据
func (pow *ProofOfWork) prepareData(nonce int32)  []byte{
	data:=bytes.Join(
		[][]byte{
			IntToHex(pow.block.Version),
			pow.block.PrevBlockHash,
			pow.block.Merkleroot,
			IntToHex(pow.block.Time),
			IntToHex(pow.block.Bits),
			IntToHex(nonce)},//利用下面的函数，先计算出nonce，再进行赋值
		[]byte{},
		)
	return data
}

//类似于main函数中的开始挖矿代码
func (pow *ProofOfWork) Run()(int32,[]byte)  {
	var nonce int32
	nonce=0
	var currenthash big.Int
	var secondehash [32]byte
	for nonce< maxnonce{
		//序列化
		data:=pow.prepareData(nonce)
		//两次哈希
		firsthash:=sha256.Sum256(data)
		secondehash=sha256.Sum256(firsthash[:])
		//fmt.Printf("%x\n",secondehash)
		//最后一次哈希翻转
		//reverBytes(secondehash[:])
		currenthash.SetBytes(secondehash[:])
		//当前的哈希是否小于目标哈希
		if currenthash.Cmp(pow.target)==-1 { //当前的哈希值小于目标值
			break
		}else {
			nonce++
		}
	}
	return nonce,secondehash[:]
}

//检查是否挖矿成功
func (pow *ProofOfWork) Validate() bool  {
	var hashInt big.Int
	data:=pow.prepareData(pow.block.Nonce)
	firsthash:=sha256.Sum256(data)
	secondhash:=sha256.Sum256(firsthash[:])
	hashInt.SetBytes(secondhash[:])
	isValid:=hashInt.Cmp(pow.target)==-1
	return isValid
}

