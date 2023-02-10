package main

import (
	"bytes"
	"encoding/gob"
	"log"
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

//序列化区块
func (b *Block) Serialize()[]byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err:=encoder.Encode(b)
	if err!=nil {
		log.Panic(err)
	}
	return result.Bytes()
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

//反序列化块
func DeserializeBlock(d []byte) *Block{
	var block Block

	decoder:=gob.NewDecoder(bytes.NewBuffer(d))
	err:=decoder.Decode(&block)
	if err!=nil {
		log.Panic(err)
	}
	return &block
}
