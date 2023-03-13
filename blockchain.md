# blockchain_go

## 选择数据库

我们需要选择数据库以存储区块链，而不是放在内存中。这个项目中我们使用的是 BlotDB，因为它简单，仅提供键值对存储。

---

在 Bitcoin 中使用两个 “bucket” 来存储数据：

1.  其中一个 bucket 是 **blocks**，它存储了描述一条链中所有块的元数据
    
    
    | key  | value |
    | :-: | :-: |
    | b + 32 字节的 block hash | block index record |
    | f + 4 字节的 file number | file information record |
    | l + 4 字节的 file number | the last block file number used |
    | R + 1 字节的 boolean | 是否正在 reindex |
    | F + 1 字节的 flag name length + flag name string | 1 byte boolean: various flags that can be on or off |
    | t + 32 字节的 transaction hash | transaction index record |
2.  另一个 bucket 是 **chainstate**，存储了一条链的状态，也就是当前所有的未花费的交易输出，和一些元数据
    
    
    | key | value |
    | :-: | :-: |
    | c + 32 字节的 transaction hash | unspent transaction output record for that transaction |
    | B | 32 字节的 block hash: the block hash up to which the database represents the unspent transaction outputs |

## blockchain 的数据结构

```go
type BlockChain struct {
	tip []byte  
	db  *bolt.DB
}
```

## blockchain 方法

```go
//将区块保存到区块链中
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {		//启动区块链数据库的读写事务
		b := tx.Bucket([]byte(blocksBucket))		//根据名字（blocksbucket=blocks）返回对应的Bucket
		blockInDb := b.Get(block.Hash)		//由区块hash找到对应区块

		if blockInDb != nil {	//若该区块存在，则return，否则将区块存入bucket
			return nil
		}

		blockData := block.Serialize()		//区块序列化以便存入bucket
		err := b.Put(block.Hash, blockData)		//将区块存入bucket key=block.hash value=序列化后的区块
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))		//获取bucket中最后一个区块的hash
		lastBlockData := b.Get(lastHash)	//取出最后一个区块hash对应的区块
		lastBlock := DeserializeBlock(lastBlockData)		//反序列化区块

		if block.Height > lastBlock.Height {		//如果区块高度 > 最后一个区块的高度，则设为最后一个区块
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash		//tip保存最后一个区块的hash
		}

		return nil
	})
	if err != nil {		//无法启动数据库
		log.Panic(err)
	}
}
```

```go
//通过ID寻找交易
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()		//迭代每个区块

	for {
		block := bci.Next()		//从最后一个区块开始，以此迭代前一个区块

		for _, tx := range block.Transactions {		//取出每个区块的交易
			if bytes.Compare(tx.ID, ID) == 0 {		//如果 交易的ID== 查找ID 则return 交易
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {		//block.PrevBlockHash == 0 查找完毕
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")	//查找失败返回空交易和error
}
```

```go
//迭代区块链
func (bc *Blockchain) Iterator() *BlockchainIterator {		
	bci := &BlockchainIterator{bc.tip, bc.db}	

	return bci
}
```

```go
//返回最后一个区块的高度
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.db.View(func(tx *bolt.Tx) error {		//找出最后一个区块
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
```

```go
//通过区块hash找到并返回它
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
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
```

```go
//获取区块链上所有区块的hash
func (bc *Blockchain) GetBlockHashes() [][]byte {
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
```

```go
//通过ID寻找交易
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()		//迭代每个区块

	for {
		block := bci.Next()		//从最后一个区块开始，以此迭代前一个区块

		for _, tx := range block.Transactions {		//取出每个区块的交易
			if bytes.Compare(tx.ID, ID) == 0 {		//如果 交易的ID== 查找ID 则return 交易
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {		//block.PrevBlockHash == 0 查找完毕
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")	//查找失败返回空交易和error
}
```

## 函数

```go
func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
```

```go
// 使用创世纪块创建新的区块链
func NewBlockchain(nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		fmt.Println("No existing blockchain found. Use \"createblockchain\"	function to create one first.")
		os.Exit(1)
	}

	var tip []byte		//tip 为数据库中存储的最后一个块的哈希
	db, err := bolt.Open(dbFile, 0600, nil)		//// db = 打开的数据库文件
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {		//打开一个读写事务	db.Update(...)
		b := tx.Bucket([]byte(blocksBucket))		//获取数据库中区块的 bucket	tx.Bucket() 按名称检索存储桶
		tip = b.Get([]byte("l"))		//区块的 Bucket 存在，b.Get([]byte("1")) 会返回数据库中 key=1 对应的值，即最后一个区块的hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}		// 

	return &bc
}
```



```go
//创建区块链的数据库
func CreateBlockchain(address, nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte		//tip 为数据库中存储的最后一个块的哈希

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)		//创建一个新的coinbase
	genesis := NewGenesisBlock(cbtx)		//创建新的创世块

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))		//创建一个新的Bucket 命名为 blocks
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())		//新桶中放入 创世块的hash（作为key）和创世块（作为value）
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)		//在key=l处，放入创世块的hash
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash		//tip 记录桶内最后一个区块的hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}		

	return &bc
}
```