package main

import "fmt"

func TestPow()  {
	block:=&Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		1418755780,
		404454260,
		0,
		[]*Transation{},
		0,
	}
	pow:=NewProofOfWork(block)
	nonce,_:=pow.Run()
	block.Nonce=nonce
	fmt.Println("POW",pow.Validate())
}


func TestcreateMerkleTreeRoot()  {
	block:=&Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		1418755780,
		404454260,
		0,
		[]*Transation{},
		0,
	}
	txin1:=TXInput{[]byte{},-1,nil,nil}
	txout1:=NewTXOutput(subsidy,"first")
	tx1:=Transation{nil,[]TXInput{txin1},[]TXOutput{*txout1}}

	txin2:=TXInput{[]byte{},-1,nil,nil}
	txout2:=NewTXOutput(subsidy,"second")
	tx2:=Transation{nil,[]TXInput{txin2},[]TXOutput{*txout2}}

	var Transations []*Transation
	Transations=append(Transations,&tx1,&tx2)
	block.createMerklTreeRoot(Transations)
	fmt.Printf("%x\n",block.Merkleroot)
}
func TestNewSerialize()  {
	block:=&Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		1418755780,
		404454260,
		0,
		[]*Transation{},
		0,
	}
	block.Deserialize(block.Serialize()).String()
}

//测试添加区块
func TestBoltDB()  {
	blockchain:=NewBlockchain("13waVMhErsKBGG1kxhp3GhW39c1mArd81E")  //初始化区块链
	blockchain.MineBlock([]*Transation{}) //添加区块链
	blockchain.MineBlock([]*Transation{})
	blockchain.printBlockchain()
}