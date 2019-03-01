package main

import "testing"

func TestWallets_GetAddress(t *testing.T) {
	var cli=CLI{}
	cli.listAddress()
}

func Test_Send(t *testing.T){
	var cli=CLI{}
	cli.send("13waVMhErsKBGG1kxhp3GhW39c1mArd81E", "1hevV2RnWGnjH5SQYs9vhoa5heki88HW2", 20)
}