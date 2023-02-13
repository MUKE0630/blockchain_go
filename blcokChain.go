package main

import (
	"bolt/bolt"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"bytes"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 12/2/2023 Chancellor on brink of second bailout for banks"

// 区块链保持一个有序快。
type BlockChain struct {
	tip []byte
	db  *bolt.DB
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
func (bc *BlockChain) FindUnSpentTranscations(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				//检查是否双花
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}

					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
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
func (bc *BlockChain) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	unspentTranscations := bc.FindUnSpentTranscations(pubKeyHash)

	for _, tx := range unspentTranscations {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

//通过ID找到交易
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

// 查找并返回未花费输出以参考输入
func (bc *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnSpentTranscations(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outidx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
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


func (bc *BlockChain) Iterator() *BlockChainIterator {
	bci := &BlockChainIterator{bc.tip, bc.db}
	return bci
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

	cbtx:=NewCoinbaseTX(address,genesisCoinbaseData)
	genesis:=NewGenesisBlock(cbtx)

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
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

//交易签名
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

//验证签名
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}