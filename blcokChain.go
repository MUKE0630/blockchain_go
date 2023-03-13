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
	tip []byte		//tip 为数据库中存储的最后一个块的哈希
	db  *bolt.DB
}

//返回最后一个区块的height
func (bc *BlockChain) GetBestHeight() int {
	var lastBlock Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

//通过hash寻找一个区块并返回
func (bc *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

//返回链上所有块的hash
func (bc *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// 用提供的交易挖出一个新的块
func (bc *BlockChain) MineBlock(transactions []*Transaction) *Block{
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("1"))

		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Height

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	NewBlock := NewBlock(transactions, lastHash,lastHeight+1)

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

	if err !=nil{
		log.Panic(err)
	}

	return NewBlock
}

// 返回包含未解决输出的交易列表
func (bc *BlockChain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// 这个输出是否被花费过？
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
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
func NewBlockChain(nodeID string) *BlockChain {
	if dbExists() == false {		//查看是否有数据库文件
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte		//tip 为数据库中存储的最后一个块的哈希
	db, err := bolt.Open(dbFile, 0600, nil)		// db = 打开的数据库文件
	if err != nil {		// 文件读取异常处理
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))		
		tip = b.Get([]byte("1"))		//b.Get([]byte("1")) 会返回数据库中 key=1 对应的值，即最后一个区块的hash

		return nil
	})

	if err != nil {		
		log.Panic(err)
	}

	bc := BlockChain{tip, db}

	return &bc
}

//将区块保存到链上
func (bc *BlockChain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
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
	if tx.IsCoinbase() {
		return true
	}
	
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