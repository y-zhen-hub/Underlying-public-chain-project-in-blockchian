package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

const dbFile  = "blockchin.db" //定义数据库存储位置，命名 dbFile
const blockBucket  = "blocks"  //定义一个桶,Buckets是bolt数据库中存放key/value对的地方
const genesisData ="jonson blockchian"

type BlockChain struct {
	tip []byte //定义最近一个区块的hash值
	db *bolt.DB
}


type BlockChainIterator struct {  //区块循环迭代
	currenthash []byte
	db *bolt.DB
}

//创建一个区块链
func NewBlockchain(address string)  *BlockChain{
	var tip []byte //
	db,err:=bolt.Open(dbFile,0600,nil) ////创建bolt数据库本地文件
	if err!=nil{
		log.Panic(err)
	}
	err=db.Update(func(tx *bolt.Tx) error {  //更新数据库
		b:=tx.Bucket([]byte(blockBucket))//存储桶
		if b==nil{
			fmt.Println("区块链不存在，创建一个新的区块链")
			transaction :=NewCoinbaseTX(address,genesisData)//传递一个交易的哈希
			genesis:=NewGensisBlock([]*Transation{transaction})//创建一个创世区块
			b,err:=tx.CreateBucket([]byte(blockBucket))
			if err!=nil{
				log.Panic(err)
			}
			err=b.Put(genesis.Hash,genesis.Serialize())
			if err!=nil{
				log.Panic(err)
			}
			err = b.Put([]byte("l"),genesis.Hash)
			tip=genesis.Hash
		}else {
			tip=b.Get([]byte("l")) //根据key,获得对应的value值
		}
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
	bc:=BlockChain{tip,db}
	set:=UTXOSet{&bc}
	set.Reindex()
	return &bc
}

//添加新的区块
func (bc *BlockChain)MineBlock(transations []*Transation)  {
	for _,tx:=range transations{
		if bc.VerifyTransation(tx)!=true{
			log.Panic("验证失败")
		}else {
			fmt.Println("验证成功")
		}
	}
	var lasthash []byte
	var lastheight int32
	err:=bc.db.View(func(tx *bolt.Tx) error {  //读取数据
		b:=tx.Bucket([]byte(blockBucket))
		lasthash=b.Get([]byte("l")) //获得最长哈希值对应的区块链
		blockdata:=b.Get(lasthash)  //获得最后一个哈希值对应的序列化数据---区块链
		block:=Deserialize(blockdata) //序列化区块链数据
		lastheight=block.Height
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
	newBlock:=NewBlock(transations,lasthash,lastheight+1)  //获取最后一个哈希值，为创建新的区块做准备

	err=bc.db.Update(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(blockBucket)) //获得一个桶
		err:=b.Put(newBlock.Hash,newBlock.Serialize())  //将数据添加到区块链中
		if err!=nil{
			log.Panic(err)
		}
		err=b.Put([]byte("l"),newBlock.Hash) //将“l”与新区块的哈希绑定
		if err!=nil {
			log.Panic(err)
		}
		bc.tip=newBlock.Hash
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
}

//将数据循环添加到数据库中，并返回结构体
func (bc *BlockChain) iterator() *BlockChainIterator {  //新建实例
	bci:=&BlockChainIterator{bc.tip,bc.db}
	return bci
}

//反序列化，返回区块的数据
func  Deserialize(d []byte) *Block {
	var block  Block
	decode:=gob.NewDecoder(bytes.NewReader(d))  //读操作
	err:=decode.Decode(&block)  //拿到结构体
	if err!=nil{
		log.Panic(err)
	}
	return &block
}

//查找前一个区块的哈希值，返回前一个区块，为打印区块链的信息做准备
func (i *BlockChainIterator)Next() *Block{  //寻找前一个哈希值
	var block *Block
	err:=i.db.View(func(tx *bolt.Tx) error {  //查找数据库中的值
		b:=tx.Bucket([]byte(blockBucket))
		deblock:=b.Get(i.currenthash)
		block= Deserialize(deblock)

		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
	i.currenthash=block.PrevBlockHash
	return block
}

//打印区块链信息
func (bc *BlockChain) printBlockchain() {
	bci:=bc.iterator()
	for{
		block:=bci.Next()
		block.String()
		fmt.Println()
		if len(block.PrevBlockHash)==0{
			break
		}
	}

}

//遍历所有的区块---遍历所有区块中的交易—遍历所有交易的未花费输出
func (bc *BlockChain)FindUnspentTransations(pubkeyhash []byte)  []Transation{
	var unspentTXs []Transation  //所有未花费的交易,该交易中一定有一笔输出是未花费的
	spendTXOs:=make(map[string][]int)  //string 交易哈希值--->[]int 输出的序号 存储已经花费的交易
	bci:=bc.iterator()
	for{  //从后往前进行遍历区块
		block:=bci.Next()   //查找前一个区块

		for _,tx := range block.Transations{  //tx对应的是每一个交易，
			txID := hex.EncodeToString(tx.ID)  //ID对应为交易哈希值，将其转化为字符串类型

		output:
			for outIdx, out :=range tx.Vout{  //遍历交易中的每笔输出，序号+输出
				if spendTXOs[txID]!=nil{ //若输出不为空--代表这笔交易肯定有已经花费的输出，但不能判断是哪些输出已经完全花费
					for _,spentOut:=range spendTXOs[txID]{ //遍历所有已经花费输出切片
						if spentOut == outIdx{  //若已经花费的输出序号==输出序号，说明该输出是已经花费的
							continue output   //重新进行output循环
						}
					}
				}
				if out.CanBeUnlockedWith(pubkeyhash){  //交易3满足条件，每笔输出可以解锁地址，说明确实是输出给Bob的，也就是说这笔交易是未花费的
					unspentTXs = append(unspentTXs,*tx)  //将当前的输出进行存储
				}
			}
			if tx.IsCoinBase()==false{  //不是矿工的输出
				for _,in :=range tx.Vin{  //遍历输入，
					if in.CanUnlockOutputWith(pubkeyhash){ //输入应该是Bob的输入
						inTxId :=hex.EncodeToString(in.TXid) //转换成字符串
						spendTXOs[inTxId]=append(spendTXOs[inTxId],in.VoutIndex) //将该输入存储
					}
				}
			}
		}
		if len(block.PrevBlockHash)==0{    //终止循环
			break
		}
	}
	return  unspentTXs
}

//根据公钥哈希找到相应地址的未花费，并存储
func (bc *BlockChain) FindUTXO(pubkeyhash  []byte)  []TXOutput{
	var UTXOs []TXOutput
	unspendTransations :=bc.FindUnspentTransations(pubkeyhash)//调用函数，接收所有包含未花费的输出
	for _,tx :=range unspendTransations{//遍历所有的未花费交易
		for _,out :=range tx.Vout{//遍历所有的输出
			if out.CanBeUnlockedWith(pubkeyhash){//找到是Bob的输出
				UTXOs =append(UTXOs,out)
			}
		}
	}
	return UTXOs
}

//账户中可以花费的输出--可以花费的具体金额
func (bc *BlockChain)FindSpendableOutputs (pubkeyhash []byte, amount int) (int,map[string][]int){//账户中是否有足够金额进行转账
	unspentOutputs :=make(map[string][]int) //为后续返回值做准备，未花费的输出
	unspentTXs :=bc.FindUnspentTransations(pubkeyhash) //获得未花费的交易
	accumulated :=0 //将要输出的值

	Work:
		for _,tx :=range unspentTXs{ //遍历所有的交易
			txID :=hex.EncodeToString(tx.ID) //将交易哈希值转换成字符串类型的

			for outIDx,out :=range tx.Vout{ //遍历所有的输出
				if out.CanBeUnlockedWith(pubkeyhash)&&accumulated<amount{ //如果确实是Bob本人输出的，并且输出金额小于账户金额
					accumulated+=out.Value  //将要进行输出的金额相加
					unspentOutputs[txID]=append(unspentOutputs[txID],outIDx) //添加到待返回的映射中
					if accumulated>=amount{  //如果不足金额的话，会直接跳出循环
						break Work
					}
				}
			}
		}

	return accumulated,unspentOutputs
}

//添加数字签名  两个参数，一个是交易指针，另外一个是私钥结构体
func (bc *BlockChain) SignTransation (tx *Transation, prikey ecdsa.PrivateKey)  {
	prevTXs:=make(map[string]Transation) //构建前一个交易ID--哈希与交易结构体的映射
	for _,vin :=range tx.Vin{
		prevTX,err:=bc.FindTransationById(vin.TXid)  //检查区块中是否有该交易哈希（ID），ID是前一笔交易的输出
		if err !=nil{
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)]=prevTX  //将ID转换成字符串，完成映射
	}
	tx.Sign(prikey,prevTXs)
}

//通过ID查找对应的输入  ---已检查
func (bc *BlockChain) FindTransationById  (ID []byte) (Transation, error){
	bci :=bc.iterator() //遍历区块链中所有的ID，得到*BlockChainIterator结构体，包括当前哈希值和区块链数据库
	for{
		block :=bci.Next()  //查找前一个区块，返回的区块是个结构体
			for _,tx :=range block.Transations{
				if bytes.Compare(tx.ID,ID)==0{ //对输入的ID进行对比，这里的ID是哈希
					return *tx,nil  //返回交易结构体
				}
			}
			if len(block.PrevBlockHash)==0{ //当到达创世区块的时候就会终止
				break
			}
	}
	//若循环完成后，还未找到，则返回错误信息
	return Transation{},errors.New("transaction is not find")
}

//验证交易以及验证交易是否有效
func (bc *BlockChain) VerifyTransation (tx *Transation) bool  {
	prevTXs :=make(map[string]Transation)

	for _,vin :=range tx.Vin{
		prevTX,err:=bc.FindTransationById(vin.TXid)
		if err!=nil{
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)]=prevTX
	}
	return tx.Verify(prevTXs)
}

//找到所有的未花费输出
func (bc *BlockChain) FindALLUTXO() map[string]TXOutputs {
	UTXO :=make(map[string]TXOutputs) //所有未花费输出结构体
	spentTXs :=make(map[string][]int) //已花费的输出

	bci :=bc.iterator() //循环整个区块链
	for {
		block:=bci.Next()
			for _,tx :=range block.Transations{ //循环遍历区块中所有的交易
				txID :=hex.EncodeToString(tx.ID)  //将交易哈希解析成字符串

			Outputs:
				for outIdx,out:= range tx.Vout{ //遍历交易中所有的输出

					if spentTXs[txID]!=nil{  //如果交易ID在已花费中

						for _,spendOutIds:=range spentTXs[txID]{ //遍历该交易中所有的已花费

							if spendOutIds ==outIdx{ //若匹配成功，则表示该输出是已经花费的
								continue Outputs
							}
						}
					}
					outs:=UTXO[txID] //若不是已经花费的，则进行存储
					outs.Outputs=append(outs.Outputs,out)
					UTXO[txID]=outs
				}

				if tx.IsCoinBase()==false{
					for _,in:=range tx.Vin{
						inTXID:=hex.EncodeToString(in.TXid) //将输入引用的前一笔交易哈希解析为字符串
						spentTXs[inTXID]=append(spentTXs[inTXID],in.VoutIndex)  //将前一交易输出存储到已花费中
					}
				}
			}
			if len(block.PrevBlockHash)==0{ //遍历到创世块后，就停止遍历
				break
			}
	}
	return UTXO
}

//获得当前区块链的高度
func (bc *BlockChain) GetBestHeight() int32  {
	var lastBlock Block

	err:=bc.db.View(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(blockBucket))
		lastHash:=b.Get([]byte("l"))
		blockdata:=b.Get(lastHash)
		lastBlock=*Deserialize(blockdata)
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
	return lastBlock.Height
}

func (bc *BlockChain) Getblockhash() [][]byte {
	var blocks [][]byte
	bci:=bc.iterator()
	for{
		block:=bci.Next()
		blocks=append(blocks,block.Hash)
		if len(block.PrevBlockHash)==0{
			break
		}
	}
	return blocks

}

func (bc *BlockChain) GetBlock(blockHash []byte) (Block,error) {
	var block Block

	err:=bc.db.Update(func(tx *bolt.Tx) error { //查询数据库
		b:=tx.Bucket([]byte(blockBucket)) //获得篮子
		blockData:=b.Get(blockHash)  //根据哈希获得对应的序列化的数据

		if blockData==nil{
			return errors.New("BLock is not Fund")
		}
		block=*Deserialize(blockData)
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
	return block,nil
}

func (bc *BlockChain) AddBlock(block *Block) {
	err:=bc.db.View(func(tx *bolt.Tx) error {//更新数据库
		b:=tx.Bucket([]byte(blockBucket))//获得篮子
		blockIndb:=b.Get(block.Hash) //获得哈希对应的区块序列化数据
		if blockIndb!=nil{
			return nil
		}
		blockData:=block.Serialize()  //对当前区块数据进行序列化
		err:=b.Put(block.Hash,blockData) //将对新的区块数据天骄
		if err!=nil{
			log.Panic(err)
		}
		lasthash:=b.Get([]byte("l"))//获得数据库中最新的区块哈希
		lastBlockdata:=b.Get(lasthash)//根据区块哈希获得数据库中最新的区块信息
		lastblock:=Deserialize(lastBlockdata) //反序列化数据库中区块信息

		if block.Height>lastblock.Height{
			err=b.Put([]byte("l"),block.Hash) //若外部节点区块链高度高于本地数据中的，则将外部最新区块哈希添加到数据库中
			if err!=nil{
				log.Panic(err)
			}
			bc.tip=block.Hash  //将数据库中最新区块哈希定义为外部节点最新哈希值
		}
		return nil
	})
	if err!=nil {
		log.Panic(err)
	}
}