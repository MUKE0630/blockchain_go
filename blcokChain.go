package main

import(
	"fmt"
	"log"

	"bolt/bolt"
)

const dbFile="blockchain.db"
const blocksBucket="blocks"
 
//区块链保持一个有序快。
type BlockChain struct {
	tip []byte
	db *bolt.DB
}


//用于对区块链块进行迭代
type BlockChainIterator struct{
	currentHash []byte
	db *bolt.DB
}




//保存数据到链上
func (bc *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err:=bc.db.View(func(tx *bolt.Tx) error{
		b:=tx.Bucket([]byte(blocksBucket))
		lastHash=b.Get([]byte("1"))
		return nil
	})

	if err!=nil {
		log.Panic(err)
	}

	newBlock:=NewBlock(data,lastHash)

	err=bc.db.Update(func(tx *bolt.Tx)error{
		b:=tx.Bucket([]byte(blocksBucket))
		err:=b.Put(newBlock.Hash,newBlock.Serialize())
		if err!=nil {
			log.Panic(err)
		}

		err=b.Put([]byte("1"),newBlock.Hash)
		if err!=nil {
			log.Panic(err)
		}

		bc.tip=newBlock.Hash
		return nil
	})
}

func (bc *BlockChain) Iterator()*BlockChainIterator{
	bci:=&BlockChainIterator{bc.tip,bc.db}
	return bci
}

//返回从tip开始的下一个块
func (i *BlockChainIterator)Next()*Block{
	var block *Block

	err:=i.db.View(func (tx*bolt.Tx) error {
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

//使用创世纪块创建新的区块链
func NewBlockChain() *BlockChain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			fmt.Println("No existing blockchain found. Creating a new one...")
			genesis := NewGenesisBlock()

			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{tip, db}

	return &bc
}
