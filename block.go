package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

// 精简后的区块头.
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height int
}

// 序列化区块
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// 返回块中的交易哈希
func (b *Block) HashTranscations() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

// 创建新块并返回
func NewBlock(transaction []*Transaction, prevBlockHash []byte,height int) *Block {
	block := &Block{time.Now().Unix(), transaction, prevBlockHash, []byte{}, 0,height}
	pow := NewProofOfWork(block)
	nonce, hash := pow.RunProofOfWork()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// 创建新的创世纪块并返回
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{},0)
}

// 反序列化块
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewBuffer(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
