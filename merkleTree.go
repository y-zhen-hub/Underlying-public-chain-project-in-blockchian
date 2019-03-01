package main

import (
	"crypto/sha256"
)

type MerkleTree struct{ //存储根节点
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left *MerkleNode
	Right *MerkleNode
	Data []byte
}


//根据子节点，生成父节点，如果已经是叶节点，则存储，若是中间节点，则进行拼接并进行哈希运算，进行存储
func NewMerkleNode(left,right *MerkleNode,data []byte)  *MerkleNode {
	mnode:=MerkleNode{}//构建一个新的节点
	if left==nil &&right==nil {
		mnode.Data=data
	}else {
		prehashes:=append(left.Data,right.Data...)
		firsthash:=sha256.Sum256(prehashes)
		secondehash:=sha256.Sum256(firsthash[:])
		mnode.Data=secondehash[:]
	}
	mnode.Left=left
	mnode.Right=right
	return &mnode
}

//构建默克尔树---所有的数据是一个二维切片  //将所有叶子节点都压缩到切片中
func NewMerkleTree(data [][] byte) *MerkleTree {
	var nodes []MerkleNode//新建节点切片
	for _,datum:=range data  {  //拿到叶子节点的哈希值
		node:=NewMerkleNode(nil,nil,datum)//将每个切片变成节点
		nodes=append(nodes,*node)//将叶子节点添加到切片后面
	}

	j:=0
	//两两拼接
	for nSize:=len(data);nSize>1;nSize=(nSize+1)/2{
		for i:=0;i<nSize;i+=2{
			i2:=min(i+1,nSize-1)//解决奇偶个数的问题
			node:=NewMerkleNode(&nodes[j+i],&nodes[j+i2],nil)
			nodes=append(nodes,*node)
		}
		j+=nSize
	}
	mTree:=MerkleTree{&(nodes[len(nodes)-1])}
	return &mTree
}
//func main()  {
//	data1,_:=hex.DecodeString("6b6a4236fb06fead0f1bd7fc4f4de123796eb51675fb55dc18c33fe12e33169d")
//	data2,_:=hex.DecodeString("2af6b6f6bc6e613049637e32b1809dd767c72f912fef2b978992c6408483d77e")
//	data3,_:=hex.DecodeString("6d76d15213c11fcbf4cc7e880f34c35dae43f8081ef30c6901f513ce41374583")
//	data4,_:=hex.DecodeString("08c3b50053b010542dca85594af182f8fcf2f0d2bfe8a806e9494e4792222ad2")
//	data5,_:=hex.DecodeString("612d035670b7b9dad50f987dfa000a5324ecb3e08745cfefa10a4cefc5544553")
//
//	reverBytes(data1)
//	reverBytes(data2)
//	reverBytes(data3)
//	reverBytes(data4)
//	reverBytes(data5)
//
//	nodes:=[][]byte{
//		data1,
//		data2,
//		data3,
//		data4,
//		data5,
//	}
//	merkleroot:=NewMerkleTree(nodes)
//	reverBytes(merkleroot.RootNode.Data)
//	fmt.Printf("%x\n",merkleroot.RootNode.Data)
//}