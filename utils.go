package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

//转化成字节数组--小端
func IntToHex(num int32)  []byte{
	buff:=new(bytes.Buffer)  //buffer是一个切片的缓冲器，可以读写文件
	//通过Write接口可以将data参数里面包含的数据写入到buffer中，通过LittleEndian将uintx类型反序列化到buf中。
	err:=binary.Write(buff,binary.LittleEndian,num)//转化为字节数组,以小端形式存储
	if err!=nil{
		log.Panic(err)
	}
	return buff.Bytes()
}
//转换成大端
func IntToHex2(num int32)  []byte{
	buff:=new(bytes.Buffer)  //buffer是一个切片的缓冲器，可以读写文件
	//通过Write接口可以将data参数里面包含的数据写入到buffer中，通过LittleEndian将uintx类型反序列化到buf中。
	err:=binary.Write(buff,binary.BigEndian,num)//转化为字节数组,以小端形式存储
	if err!=nil{
		log.Panic(err)
	}
	return buff.Bytes()
}
//大小端的转换
func reverBytes(data []byte) []byte{
	for i,j:=0,len(data)-1;i<j;i,j=i+1,j-1{
		data[i],data[j]=data[j],data[i]
	}
	return data
}

//计算两个数的最小值
func min(a,b int )int  {
	if(a>b){
		return b
	}else {
		return a
	}
}