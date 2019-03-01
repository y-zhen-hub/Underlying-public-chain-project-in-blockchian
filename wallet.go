package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/ripemd160"
)

const version  = byte(0x00)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	Publickey []byte
}
//生车成新的公钥私钥---钱包
func NewWallet() *Wallet {
	private,public:=newKeypair()
	wallet :=Wallet{private,public}
	return &wallet
}
//产生公私钥
func newKeypair() (ecdsa.PrivateKey,[]byte) {
	//生成椭圆曲线，secp256r1曲线，比特币中的曲线是secp256k1
	curve:=elliptic.P256()
	//产生私钥
	private,err:=ecdsa.GenerateKey(curve,rand.Reader)
	if err!=nil{
		fmt.Println("error")
	}
	//生成公钥，private.PublicKey.X.Bytes()是坐标轴上x的值，另外一个是y的值
	publickey:=append(private.PublicKey.X.Bytes(),private.PublicKey.Y.Bytes()...)
	return *private,publickey
}



//对公钥进行哈希运算
func HashPubkey(pubkey []byte)  []byte{
	//生成公钥的哈希值
	pubkeyHash256 :=sha256.Sum256(pubkey)
	//生成RIPEMD160算法
	RIPEMD160Hasher:=ripemd160.New()
	//进行RIPEMD160算法，生成公钥的哈希值
	_,err:=RIPEMD160Hasher.Write(pubkeyHash256[:])
	if err!=nil{
		fmt.Println("error")
	}
	pubRIPEMD160:=RIPEMD160Hasher.Sum(nil)

	return pubRIPEMD160
}
//获得checksum
func CheckSum(payload []byte)  []byte{
	firstSHA :=sha256.Sum256(payload)
	secondSHA:=sha256.Sum256(firstSHA[:])
	//chechsum是前四个字节
	checksum :=secondSHA[:4]
	return checksum
}

//将之前的函数拆分进行封装，获得地址--私钥-->公钥-->地址
func (w Wallet)GetAddress() []byte {
	//获得公钥哈希
	publickHash :=HashPubkey(w.Publickey)
	//将版本号和公钥哈希拼接
	versionPayload :=append([]byte{version},publickHash...)
	//获得检查值
	check :=CheckSum(versionPayload)
	//将前面的三个数据进行拼接
	fullPayload:=append(versionPayload,check...)
	//返回地址
	address:=Base58Encode(fullPayload)
	return address

}
//验证地址是否有效
func ValidAddress(address []byte) bool {
	pubkeyHash :=Base58Decode([]byte(address))
	//后四位
	actualCheckSum :=pubkeyHash[len(pubkeyHash)-4:]
	//中间的哈希值，除去版本号和检查值
	publickeyHash :=pubkeyHash[1:len(pubkeyHash)-4]
	targetCheckSum :=CheckSum(append([]byte{0x00},publickeyHash...))
	return bytes.Compare(actualCheckSum,targetCheckSum)==0
}