package main

import (
	"bytes"
	"math/big"
)

//base58编码   给字节数组赋值为“” 中的内容
var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
func Base58Encode(input []byte) []byte {

	//用来接收余数
	var result []byte

	x:=big.NewInt(0).SetBytes(input)  //新建大整数,先初始为0，后续设置为input
	base:=big.NewInt(int64(len(b58Alphabet)))//表示有多少位--转换成大整数
	zero:=big.NewInt(0)  //零的大整数

	//构建大整数指针
	mod := &big.Int{}
	//循环，不停对x取余，商为58
	for x.Cmp(zero) !=0{  //Cmp函数比较x与zero的值，若果x小则结果为-1，如果大则为1，如果相等则为0
		x.DivMod(x,base,mod) //对58长度取余，x/base,mod存储最后取余的值，结果是商
		//将余数存储在数组中
		result=append(result,b58Alphabet[mod.Int64()]) //将余数存储在数组中

	}
	reverseBytes(result)
	for _,b:=range input{
		if b==0x00{
			//一个切片添加到另一个切片中需要用“...
			//如果前面的是0就需要在当前位置添加1
			result=append([]byte{b58Alphabet[0]},result...)
		}else{
			break
		}
	}
	return result
}
//反转所有的余数值
func reverseBytes(data [] byte)  [] byte{
	for i,j:=0,len(data)-1;i<j ;i,j=i+1,j-1  {
		data[i],data[j]=data[j],data[i]
	}
	return data
}

//解码
func Base58Decode(input [] byte)  [] byte{
	result :=big.NewInt(0)
	zeroBytes :=0  //前面有多少个1
	for _,b:=range input{
		if b=='1'{
			zeroBytes++
		}else {
			break
		}
	}
	//除去前面的1
	payload := input[zeroBytes:]  //存储zeroBytes后面的数据
	for _,b:=range payload {
		//将b的值对应到b58Alphabet上的位置---反推出余数
		charIndex :=bytes.IndexByte(b58Alphabet,b)
		//商乘以58，获得相应的数
		result.Mul(result,big.NewInt(58))
		//上面的数值加上余数就是被除数
		result.Add(result,big.NewInt(int64(charIndex)))

	}
	decoded :=result.Bytes()
	//处理0
	decoded=append(bytes.Repeat([]byte{0x00},zeroBytes),decoded...)
	return decoded
}
