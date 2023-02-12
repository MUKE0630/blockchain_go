package main

import (
	"bolt/bolt"
	"encoding/hex"
	//"errors"
	"fmt"
	"log"
	"os"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 12/2/2023 Chancellor on brink of second bailout for banks"

// 区块链保持一个有序快。
type BlockChain struct {
	tip []byte
	db  *bolt.DB
}

// 用于对区块链块进行迭代
type BlockChainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// 用提供的交易挖出一个新的块
func (bc *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("1"))

		return nil
	})

	
	if err != nil {
		log.Panic(err)
	}

	NewBlock := NewBlock(transactions, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(NewBlock.Hash, NewBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("1"), NewBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = NewBlock.Hash

		return nil
	})
}

// 返回包含未解决输出的交易列表
func (bc *BlockChain) FindUnSpentTranscations(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				//检查是否双花
				if spentTXs[txID] != nil {
					for _, spentOut := range spentTXs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}

					}
				}

				if out.CanBeUnlockWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXs[inTxID] = append(spentTXs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// 查找并返回所有未解决的交易输出
func (bc *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTranscations := bc.FindUnSpentTranscations(address)

	for _, tx := range unspentTranscations {
		for _, out := range tx.Vout {
			if out.CanBeUnlockWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// 查找并返回未花费输出以参考输入
func (bc *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnSpentTranscations(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outidx, out := range tx.Vout {
			if out.CanBeUnlockWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outidx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

// 保存数据到链上
// func (bc *BlockChain) AddBlock(data string) {
// 	var lastHash []byte

// 	err := bc.db.View(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte(blocksBucket))
// 		lastHash = b.Get([]byte("1"))
// 		return nil
// 	})

// 	if err != nil {
// 		log.Panic(err)
// 	}

// 	newBlock := NewBlock(data, lastHash)

// 	err = bc.db.Update(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte(blocksBucket))
// 		err := b.Put(newBlock.Hash, newBlock.Serialize())
// 		if err != nil {
// 			log.Panic(err)
// 		}

// 		err = b.Put([]byte("1"), newBlock.Hash)
// 		if err != nil {
// 			log.Panic(err)
// 		}

// 		bc.tip = newBlock.Hash
// 		return nil
// 	})
// }

func (bc *BlockChain) Iterator() *BlockChainIterator {
	bci := &BlockChainIterator{bc.tip, bc.db}
	return bci
}

// 返回从tip开始的下一个块
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

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsExist(err) {
		return false
	}
	return true
}

// 使用创世纪块创建新的区块链
func NewBlockChain(address string) *BlockChain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("1"))

		return nil
	})

	// 	if b == nil {
	// 		fmt.Println("No existing blockchain found. Creating a new one...")
	// 		genesis := NewGenesisBlock()

	// 		b, err := tx.CreateBucket([]byte(blocksBucket))
	// 		if err != nil {
	// 			log.Panic(err)
	// 		}

	// 		err = b.Put(genesis.Hash, genesis.Serialize())
	// 		if err != nil {
	// 			log.Panic(err)
	// 		}

	// 		err = b.Put([]byte("l"), genesis.Hash)
	// 		if err != nil {
	// 			log.Panic(err)
	// 		}
	// 		tip = genesis.Hash
	// 	} else {
	// 		tip = b.Get([]byte("l"))
	// 	}

	// 	return nil

	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{tip, db}

	return &bc
}

// 创建一个新的区块链DB
func CreateBlockchain(address string) *BlockChain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

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

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{tip, db}

	return &bc
}
