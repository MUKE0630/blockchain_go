package main

import (
	"fmt"
	"strconv"
)

func main() { 
	bc := NewBlockChain()

	bc.AddBlock("send some btc to allen")
	bc.AddBlock("send some qq to bob")

	for _, block := range bc.blocks {
		pow:=NewProofOfWork(block)
		fmt.Printf("pow:%s\n",strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
