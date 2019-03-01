package main

import (
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
)

type UTXOSet struct {  //封装对象--方便处理UTXO动作
	bchain *BlockChain
}

const utxoBucket = "chainset" //添加数据库

//将UTXO放到文件中
func (u UTXOSet) Reindex() {
	db:=u.bchain.db //拿到数据库对象
	bucketName :=[]byte(utxoBucket)  //获得桶的名字
	err1:=db.Update(func(tx *bolt.Tx) error { //更新数据库
		err2:=tx.DeleteBucket(bucketName) //删除数据库
		if err2!=nil&&err2!=bolt.ErrBucketNotFound{ //除了第一次未找到外还有错误，则输出错误
			log.Panic(err2)
		}

		_,err3:=tx.CreateBucket(bucketName) //从新创建数据库
		if err3!=nil{
			log.Panic(err3)
		}
		return nil
	})
	if err1!=nil{
		log.Panic(err1)
	}
	UTXO:=u.bchain.FindALLUTXO() //找到所有的未花费输出

	err4:=db.Update(func(tx *bolt.Tx) error {
		b:=tx.Bucket(bucketName)

		for txID,outs:=range UTXO{
			key,err5:=hex.DecodeString(txID) //将字符串解析为字节数组
			if err5!=nil{
				log.Panic(err5)
			}
			err6:=b.Put(key,outs.Serialize())//对数组进行序列化--添加切片序列化函数
			if err6!=nil{
				log.Panic(err6)
			}
		}
		return nil
	})
	if err4!=nil{
		log.Panic(err4)
	}
}

//根据公钥返回所有的未花费输出
func (u UTXOSet) FindUTXObyPubkeyHash (pubkeyhash []byte) []TXOutput {
	var UTXOs []TXOutput
	db :=u.bchain.db //读取数据库
	err:=db.View(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(utxoBucket)) //获得桶
		c:=b.Cursor() //指明位置--游标
		for k,v:=c.First();k!=nil;k,v=c.Next(){ //从第一个键值对进行循环，当交易哈希为零的时候则继续循环下一个键值对
			outs:=DeserializeOutputs(v) //v是序列化的，需要进行反序列化

			for _,out:=range outs.Outputs{ //循环所有的输出
				if out.CanBeUnlockedWith(pubkeyhash){ //若未花费输出可以被该公钥拥有者解锁，则进行存储
					UTXOs=append(UTXOs,out)
				}
			}
		}
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
	return UTXOs
}

//有新的未花费输出时，更新未花费列表---在生成一个新的区块或接收一个新的区块的时候会调用该函数
//有两种情况需要进行考虑，一种是当有新的区块添加时，旧时未花费输出是否已经花费，这里需要进行判断，
//第二种是新区块产生时，最新的未花费输出需要添加到未花费列表中
func (u UTXOSet) update(block *Block)  {
	db:=u.bchain.db //获得数据库
	err:=db.Update(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(utxoBucket))//获得桶

		for _,tx:=range block.Transations{
			if tx.IsCoinBase()==false{
				for _,vin:=range tx.Vin{ //遍历所有的输入
					updateouts:=TXOutputs{} //存储所有的未花费输出
					outsbytes:=b.Get(vin.TXid)//获得当前交易引用的前一笔交易的所有未花费输出--是序列化的
					outs:=DeserializeOutputs(outsbytes)//反序列化操作

					for outIdx,out:=range outs.Outputs { //将所有的未花费输出进行键值对分离
						if outIdx!=vin.VoutIndex{ //未花费输出的索引值和当前交易引用的输出索引值不相等时，则表示该未花费输出仍旧是未花费的
							updateouts.Outputs=append(updateouts.Outputs,out)//将未花费输出存储
						}
					}
					if len(updateouts.Outputs)==0{  //若未花费输出的切片中不存在数据，则表示已经花费，将该交易引用的交易哈希删除
						err:=b.Delete(vin.TXid)
						if err!=nil{
							log.Panic(err)
						}
					}else {
						err:=b.Put(vin.TXid,updateouts.Serialize())//将交易对应的所有的未花费输出按键值对存储
						if err!=nil {
							log.Panic(err)
						}
					}
				}
			}
			newOutputs :=TXOutputs{}
			for _,out:=range tx.Vout{  //将最新的未花费输出存储
				newOutputs.Outputs=append(newOutputs.Outputs,out)
			}
			err:=b.Put(tx.ID,newOutputs.Serialize())  //这里的tx.ID和vin.VoutIndex代表的是同一个哈希值
			if err!=nil {
				log.Panic(err)
			}
		}
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
}