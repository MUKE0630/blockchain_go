package main

import (
	"log"

	"bolt/bolt"
)

// BlockchainIterator is used to iterate over blockchain blocks
type BlockChainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// 下一个区块
func (i *BlockChainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}