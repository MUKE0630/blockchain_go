package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

//精简后的区块头.
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce		  int
}

//设置哈希计算并计算块头哈希
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

//创建新块并返回
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{},0}
	pow:=NewProofOfWork(block)
	nonce,hash:=pow.RunProofOfWork()
	
	block.Hash=hash[:]
	block.Nonce=nonce
	return block
}

//创建新的创世纪块并返回
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
